package hash_map

import (
	"encoding/binary"
	"errors"
	mathrand "math/rand"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

type HashMap struct {
	name        string
	Tx          store_db_interface.StoreDBTransactionInterface
	Count       uint64
	Changes     map[string]*ChangesMapElement
	Committed   map[string]*CommittedMapElement
	KeyLength   int
	Deserialize func([]byte, []byte) (helpers.SerializableInterface, error)
	StoredEvent func([]byte, *CommittedMapElement)
	Indexable   bool
}

func (hashMap *HashMap) GetIndexByKey(key string) (uint64, error) {
	if !hashMap.Indexable {
		return 0, errors.New("HashMap is not Indexable")
	}

	data := hashMap.Tx.Get(hashMap.name + ":listKey:" + key)
	if data == nil {
		return 0, errors.New("Key not found")
	}

	return strconv.ParseUint(string(data), 10, 64)
}

func (hashMap *HashMap) GetKeyByIndex(index uint64) (key []byte, err error) {

	if !hashMap.Indexable {
		return nil, errors.New("HashMap is not Indexable")
	}

	if index > hashMap.Count {
		return nil, errors.New("Index exceeds count")
	}

	key = hashMap.Tx.Get(hashMap.name + ":list:" + strconv.FormatUint(index, 10))
	if key == nil {
		return nil, errors.New("Not found")
	}

	return
}

func (hashMap *HashMap) GetByIndex(index uint64) (data helpers.SerializableInterface, err error) {

	key, err := hashMap.GetKeyByIndex(index)
	if err != nil {
		return nil, err
	}

	return hashMap.Get(string(key))
}

func (hashMap *HashMap) GetRandom() (data helpers.SerializableInterface, err error) {

	if !hashMap.Indexable {
		return nil, errors.New("HashMap is not Indexable")
	}

	index := mathrand.Uint64() % hashMap.Count

	return hashMap.GetByIndex(index)
}

func (hashMap *HashMap) CloneCommitted() (err error) {

	for key, v := range hashMap.Committed {
		if v.Element != nil {
			if v.Element, err = hashMap.Deserialize([]byte(key), helpers.CloneBytes(v.Element.SerializeToBytes())); err != nil {
				return
			}
		}
	}

	return
}

func CreateNewHashMap(tx store_db_interface.StoreDBTransactionInterface, name string, keyLength int, indexable bool) (hashMap *HashMap) {

	if len(name) <= 4 {
		panic("Invalid name")
	}

	hashMap = &HashMap{
		name:      name + ":",
		Committed: make(map[string]*CommittedMapElement),
		Changes:   make(map[string]*ChangesMapElement),
		Tx:        tx,
		Count:     0,
		KeyLength: keyLength,
		Indexable: indexable,
	}

	buffer := tx.Get(hashMap.name + ":count")
	if buffer != nil {
		count, p := binary.Uvarint(buffer)
		if p <= 0 {
			panic("Error reading")
		}
		hashMap.Count = count
	}

	return
}

func (hashMap *HashMap) UnsetTx() {
	hashMap.Tx = nil
}

func (hashMap *HashMap) Get(key string) (helpers.SerializableInterface, error) {

	if len(key) != hashMap.KeyLength {
		return nil, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element, nil
	}

	var outData []byte

	if exists2 := hashMap.Committed[key]; exists2 != nil {
		if exists2.Element != nil {
			outData = helpers.CloneBytes(exists2.Element.SerializeToBytes())
		}
	} else {
		outData = hashMap.Tx.Get(hashMap.name + ":map:" + key)
	}

	var out helpers.SerializableInterface
	var err error
	if outData != nil {
		if out, err = hashMap.Deserialize([]byte(key), outData); err != nil {
			return nil, err
		}
	}
	hashMap.Changes[key] = &ChangesMapElement{out, "view"}
	return out, nil
}

func (hashMap *HashMap) Exists(key string) (bool, error) {

	if len(key) != hashMap.KeyLength {
		return false, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element != nil, nil
	}

	if exists2 := hashMap.Committed[key]; exists2 != nil {
		return exists2.Element != nil, nil
	}

	outData := hashMap.Tx.Get(hashMap.name + ":map:" + key)

	return outData != nil, nil
}

func (hashMap *HashMap) Update(key string, data helpers.SerializableInterface) error {

	if len(key) != hashMap.KeyLength {
		return errors.New("key length is invalid")
	}

	if data == nil {
		return errors.New("Data is null and it should not be")
	}

	exists := hashMap.Changes[key]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[key] = exists
	}
	exists.Status = "update"
	exists.Element = data
	return nil
}

func (hashMap *HashMap) Delete(key string) {
	exists := hashMap.Changes[key]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[key] = exists
	}
	exists.Status = "del"
	exists.Element = nil
	return
}

func (hashMap *HashMap) CommitChanges() {

	var removed []string

	for k, v := range hashMap.Changes {

		if v.Status == "del" || v.Status == "update" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommittedMapElement)
				hashMap.Committed[k] = committed
			}

			if v.Status == "del" {
				committed.Status = "del"
				committed.Stored = ""
				committed.Element = nil
				v.Status = "view"
			} else if v.Status == "update" {
				committed.Status = "update"
				committed.Stored = ""
				committed.Element = v.Element
				removed = append(removed, k)
			}

		}

	}

	for _, k := range removed {
		delete(hashMap.Changes, k)
	}

}

func (hashMap *HashMap) Rollback() {
	hashMap.Changes = make(map[string]*ChangesMapElement)
}

func (hashMap *HashMap) Reset() {
	hashMap.Committed = make(map[string]*CommittedMapElement)
}

func (hashMap *HashMap) WriteToStore() (err error) {

	if len(hashMap.Committed) == 0 {
		return
	}

	for k, v := range hashMap.Committed {
		if len(k) != hashMap.KeyLength {
			return errors.New("key length is invalid")
		}
		if v.Status == "del" {

			v.Status = "view"
			if hashMap.Tx.Exists(hashMap.name + k) {
				if err = hashMap.Tx.Delete(hashMap.name + ":map:" + k); err != nil {
					return
				}
				hashMap.Count -= 1
				if hashMap.Indexable {
					countStr := strconv.FormatUint(hashMap.Count, 10)
					if err = hashMap.Tx.Delete(hashMap.name + ":list:" + countStr); err != nil {
						return
					}
					if err = hashMap.Tx.Delete(hashMap.name + ":listKey:" + k); err != nil {
						return
					}
				}
				v.Stored = "del"
			} else {
				v.Stored = "view"
			}

		}
	}

	for k, v := range hashMap.Committed {
		if v.Status == "update" {

			if !hashMap.Tx.Exists(hashMap.name + ":map:" + k) {
				if hashMap.Indexable {
					if err = hashMap.Tx.Put(hashMap.name+":list:"+strconv.FormatUint(hashMap.Count, 10), []byte(k)); err != nil {
						return
					}
					if err = hashMap.Tx.Put(hashMap.name+":listKey:"+k, []byte(strconv.FormatUint(hashMap.Count, 10))); err != nil {
						return
					}
				}
				if hashMap.StoredEvent != nil {
					hashMap.StoredEvent([]byte(k), v)
				}
				hashMap.Count += 1
			}

			if err = hashMap.Tx.Put(hashMap.name+":map:"+k, v.Element.SerializeToBytes()); err != nil {
				return
			}
			v.Status = "view"
			v.Stored = "update"
		}

	}

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, hashMap.Count)
	if err = hashMap.Tx.Put(hashMap.name+":count", buf[:n]); err != nil {
		return
	}

	return
}
