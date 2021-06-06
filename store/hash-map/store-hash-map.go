package hash_map

import (
	"errors"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type CommittedMapElement struct {
	Data   []byte
	Status string
	Stored string
}

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

func (hashMap *HashMap) Get(key string) (out helpers.SerializableInterface, err error) {

	exists := hashMap.Changes[key]
	if exists != nil {
		out = exists.Element
		return
	}

	var outData []byte

	exists2 := hashMap.Committed[key]
	if exists2 != nil {
		outData = exists2.Data
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

func (hashMap *HashMap) Exists(key string) bool {

	exists := hashMap.Changes[key]
	if exists != nil {
		return exists.Element != nil
	}

	exists2 := hashMap.Committed[key]
	if exists2 != nil {
		return exists2.Data != nil
	}

	out := hashMap.Tx.Get(key)
	hashMap.Committed[key] = &CommittedMapElement{
		out,
		"view",
		"",
	}
	return out != nil
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

func (hashMap *HashMap) Commit() {

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
				committed.Data = nil
			} else if v.Status == "update" {
				committed.Status = "update"
				committed.Stored = ""
				committed.Data = v.Element.SerializeToBytes()
			}

			v.Status = "view"
		}

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
			if err = hashMap.Tx.Put(k, v.Data); err != nil {
				return
			}
			v.Status = "view"
			v.Stored = "update"
		}

	}

	return
}
