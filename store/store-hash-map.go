package store

import (
	"errors"
	"go.etcd.io/bbolt"
)

type CommittedMapElement struct {
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
	Committed map[string]*CommittedMapElement
	KeyLength int
}

func CreateNewHashMap(tx *bbolt.Tx, name string, keyLength int) (hashMap *HashMap) {
	hashMap = &HashMap{
		Committed: make(map[string]*CommittedMapElement),
		Changes:   make(map[string]*ChangesMapElement),
		Bucket:    tx.Bucket([]byte(name)),
		KeyLength: keyLength,
	}
	return
}

func (hashMap *HashMap) Get(key []byte) (out []byte) {
	return hashMap.get(key, true)
}

func (hashMap *HashMap) get(key []byte, includeChanges bool) (out []byte) {
	keyStr := string(key)

	if includeChanges {
		exists := hashMap.Changes[keyStr]
		if exists != nil {
			out = exists.Data
			return
		}
	}

	exists2 := hashMap.Committed[keyStr]
	if exists2 != nil {
		out = exists2.Data
		return
	}

	out = hashMap.Bucket.Get(key)
	hashMap.Committed[keyStr] = &CommittedMapElement{
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
	hashMap.Committed[keyStr] = &CommittedMapElement{
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
	exists.Status = "update"
	exists.Data = data
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

		if v.Status == "del" || v.Status == "update" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommittedMapElement)
				hashMap.Committed[k] = committed
			}

			if v.Status == "del" && committed.Status != "del" {
				committed.Status = "del"
				committed.Commit = ""
				committed.Data = nil
			} else if v.Status == "update" {
				committed.Status = "update"
				committed.Commit = ""
				committed.Data = v.Data
			}

		}

	}
	hashMap.Changes = make(map[string]*ChangesMapElement)
}

func (hashMap *HashMap) Rollback() {
	hashMap.Changes = make(map[string]*ChangesMapElement)
}

func (hashMap *HashMap) WriteToStore() (err error) {

	for k, v := range hashMap.Committed {

		if len(k) != 20 {
			errors.New("key length is invalid")
		}

		if v.Status == "del" {
			if err = hashMap.Bucket.Delete([]byte(k)); err != nil {
				return
			}
			v.Status = "view"
			v.Commit = "del"
		} else if v.Status == "update" {
			if err = hashMap.Bucket.Put([]byte(k), v.Data); err != nil {
				return
			}
			v.Commit = "update"
			v.Status = "view"
		}

	}

	return nil
}
