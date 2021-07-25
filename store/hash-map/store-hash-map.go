package hash_map

import (
	"errors"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type HashMap struct {
	name        string
	Tx          store_db_interface.StoreDBTransactionInterface
	Changes     map[string]*ChangesMapElement
	Committed   map[string]*CommittedMapElement
	KeyLength   int
	Deserialize func([]byte) (helpers.SerializableInterface, error)
}

func (hashMap *HashMap) CloneCommitted() (err error) {

	for _, v := range hashMap.Committed {
		if v.Element != nil {
			if v.Element, err = hashMap.Deserialize(helpers.CloneBytes(v.Element.SerializeToBytes())); err != nil {
				return
			}
		}
	}

	return
}

func CreateNewHashMap(tx store_db_interface.StoreDBTransactionInterface, name string, keyLength int) (hashMap *HashMap) {

	if len(name) <= 4 {
		panic("Invalid name")
	}

	hashMap = &HashMap{
		name:      name + ":",
		Committed: make(map[string]*CommittedMapElement),
		Changes:   make(map[string]*ChangesMapElement),
		Tx:        tx,
		KeyLength: keyLength,
	}
	return
}

func (hashMap *HashMap) UnsetTx() {
	hashMap.Tx = nil
}

func (hashMap *HashMap) Get(key string) (helpers.SerializableInterface, error) {

	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element, nil
	}

	var outData []byte

	if exists2 := hashMap.Committed[key]; exists2 != nil {
		if exists2.Element != nil {
			outData = helpers.CloneBytes(exists2.Element.SerializeToBytes())
		}
	} else {
		outData = hashMap.Tx.Get(hashMap.name + key)
	}

	var out helpers.SerializableInterface
	var err error
	if outData != nil {
		if out, err = hashMap.Deserialize(outData); err != nil {
			return nil, err
		}
	}
	hashMap.Changes[key] = &ChangesMapElement{out, "view"}
	return out, nil
}

func (hashMap *HashMap) Exists(key string) (bool, error) {

	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element != nil, nil
	}

	if exists2 := hashMap.Committed[key]; exists2 != nil {
		return exists2.Element != nil, nil
	}

	outData := hashMap.Tx.Get(hashMap.name + key)

	var out helpers.SerializableInterface
	var err error

	if outData != nil {
		if out, err = hashMap.Deserialize(outData); err != nil {
			return false, err
		}
	}

	hashMap.Changes[key] = &ChangesMapElement{out, "view"}
	return out != nil, nil
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

	for k, v := range hashMap.Committed {

		if len(k) != hashMap.KeyLength {
			return errors.New("key length is invalid")
		}

		if v.Status == "del" {
			if err = hashMap.Tx.DeleteForcefully(hashMap.name + k); err != nil {
				return
			}
			v.Status = "view"
			v.Stored = "del"
		} else if v.Status == "update" {
			if err = hashMap.Tx.Put(hashMap.name+k, v.Element.SerializeToBytes()); err != nil {
				return
			}
			v.Status = "view"
			v.Stored = "update"
		}

	}

	return
}
