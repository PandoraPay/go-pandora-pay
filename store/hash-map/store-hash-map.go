package hash_map

import (
	"errors"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type ChangesMapElement struct {
	Element helpers.SerializableInterface
	Status  string
}

type HashMap struct {
	Tx          store_db_interface.StoreDBTransactionInterface
	Changes     map[string]*ChangesMapElement
	Committed   map[string]*CommittedMapElement
	KeyLength   int
	Deserialize func([]byte) (helpers.SerializableInterface, error)
}

func CreateNewHashMap(tx store_db_interface.StoreDBTransactionInterface, name string, keyLength int) (hashMap *HashMap) {
	hashMap = &HashMap{
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

func (hashMap *HashMap) Get(key string) (out helpers.SerializableInterface, err error) {

	exists := hashMap.Changes[key]
	if exists != nil {
		out = exists.Element
		return
	}

	var outData []byte

	exists2 := hashMap.Committed[key]
	if exists2 != nil {
		outData = helpers.CloneBytes(exists2.Element.SerializeToBytes())
	} else {
		outData = hashMap.Tx.Get(key)
	}
	if outData != nil {
		if out, err = hashMap.Deserialize(outData); err != nil {
			return
		}
	}
	hashMap.Changes[key] = &ChangesMapElement{out, "view"}
	return
}

func (hashMap *HashMap) Exists(key string) (bool, error) {

	exists := hashMap.Changes[key]
	if exists != nil {
		return exists.Element != nil, nil
	}

	exists2 := hashMap.Committed[key]
	if exists2 != nil {
		return exists2.Element != nil, nil
	}

	outData := hashMap.Tx.Get(key)

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

func (hashMap *HashMap) Update(key string, data helpers.SerializableInterface) {
	exists := hashMap.Changes[key]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[key] = exists
	}
	exists.Status = "update"
	exists.Element = data
	return
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

func (hashMap *HashMap) WriteToStore() (err error) {

	for k, v := range hashMap.Committed {

		if len(k) != hashMap.KeyLength {
			return errors.New("key length is invalid")
		}

		if v.Status == "del" {
			if err = hashMap.Tx.DeleteForcefully(k); err != nil {
				return
			}
			v.Status = "view"
			v.Stored = "del"
		} else if v.Status == "update" {
			if err = hashMap.Tx.Put(k, v.Element.SerializeToBytes()); err != nil {
				return
			}
			v.Status = "view"
			v.Stored = "update"
		}

	}

	return
}
