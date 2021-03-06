package store

import (
	"go.etcd.io/bbolt"
)

type CommitedMapElement struct {
	Data   []byte
	Status string
	Commit string
}

type ChangesMapElement struct {
	Data   []byte
	Status string
}

type HashMap struct {
	Bucket    *bbolt.Bucket
	Changes   map[string]*ChangesMapElement
	Committed map[string]*CommitedMapElement
	KeyLength int
}

func CreateNewHashMap(tx *bbolt.Tx, name string, keyLength int) (hashMap *HashMap) {

	if tx == nil {
		panic("DB Transaction is not set")
	}

	hashMap = &HashMap{
		Committed: make(map[string]*CommitedMapElement),
		Changes:   make(map[string]*ChangesMapElement),
		Bucket:    tx.Bucket([]byte(name)),
		KeyLength: keyLength,
	}
	return
}

func (hashMap *HashMap) Get(key []byte) (out []byte) {

	keyStr := string(key)

	exists := hashMap.Changes[keyStr]
	if exists != nil {
		out = exists.Data
		return
	}

	exists2 := hashMap.Committed[keyStr]
	if exists2 != nil {
		out = exists2.Data
		return
	}

	out = hashMap.Bucket.Get(key)
	hashMap.Committed[keyStr] = &CommitedMapElement{
		out,
		"view",
		"",
	}
	return
}

func (hashMap *HashMap) Exists(key []byte) bool {
	keyStr := string(key)

	exists := hashMap.Changes[keyStr]
	if exists != nil {
		return exists.Data != nil
	}

	exists2 := hashMap.Committed[keyStr]
	if exists2 != nil {
		return exists2.Data != nil
	}

	out := hashMap.Bucket.Get(key)
	hashMap.Committed[keyStr] = &CommitedMapElement{
		out,
		"view",
		"",
	}
	return out != nil
}

func (hashMap *HashMap) Update(key []byte, data []byte) {

	keyStr := string(key)

	exists := hashMap.Changes[keyStr]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[keyStr] = exists
	}
	exists.Data = data
	exists.Status = "update"

	return
}

func (hashMap *HashMap) Delete(key []byte) {

	keyStr := string(key)

	exists := hashMap.Changes[keyStr]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[keyStr] = exists
	}
	exists.Status = "del"
	exists.Data = nil
	return
}

func (hashMap *HashMap) Commit() {
	for k, v := range hashMap.Changes {

		key := []byte(k)
		if len(key) != hashMap.KeyLength {
			panic("KeyLength is invalid")
		}

		if v.Status == "del" || v.Status == "update" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommitedMapElement)
				hashMap.Committed[k] = committed
			}

			if v.Status == "del" && committed.Status != "del" {
				committed.Status = "del"
				committed.Commit = ""
				committed.Data = nil
			} else if v.Status == "update" && committed.Status != "update" {
				committed.Status = "update"
				committed.Commit = "nil"
				committed.Data = v.Data
			}

		}

	}
	hashMap.Changes = make(map[string]*ChangesMapElement)
}

func (hashMap *HashMap) Rollback() {
	hashMap.Changes = make(map[string]*ChangesMapElement)
}

func (hashMap *HashMap) CommitToStore() {

	for k, v := range hashMap.Committed {

		key := []byte(k)
		if len(key) != hashMap.KeyLength {
			panic("KeyLength is invalid")
		}

		if v.Status == "del" {
			if err := hashMap.Bucket.Delete(key); err != nil {
				panic(err)
			}
			v.Status = "view"
			v.Commit = "del"
			v.Data = nil
		} else if v.Status == "update" {
			if err := hashMap.Bucket.Put(key, v.Data); err != nil {
				panic(err)
			}
			v.Commit = "update"
			v.Status = "view"
		}

	}

	return
}
