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
	Changes   map[[20]byte]*ChangesMapElement
	Committed map[[20]byte]*CommitedMapElement
	KeyLength int
}

func CreateNewHashMap(tx *bbolt.Tx, name string, keyLength int) (hashMap *HashMap) {

	if tx == nil {
		panic("DB Transaction is not set")
	}

	hashMap = &HashMap{
		Committed: make(map[[20]byte]*CommitedMapElement),
		Changes:   make(map[[20]byte]*ChangesMapElement),
		Bucket:    tx.Bucket([]byte(name)),
		KeyLength: keyLength,
	}
	return
}

func (hashMap *HashMap) Get(key *[20]byte) (out []byte) {
	return hashMap.get(key, true)
}

func (hashMap *HashMap) get(key *[20]byte, includeChanges bool) (out []byte) {

	if includeChanges {
		exists := hashMap.Changes[*key]
		if exists != nil {
			out = exists.Data
			return
		}
	}

	exists2 := hashMap.Committed[*key]
	if exists2 != nil {
		out = exists2.Data
		return
	}

	out = hashMap.Bucket.Get(key[:])
	hashMap.Committed[*key] = &CommitedMapElement{
		out,
		"view",
		"",
	}
	return
}

func (hashMap *HashMap) Exists(key *[20]byte) bool {

	exists := hashMap.Changes[*key]
	if exists != nil {
		return exists.Data != nil
	}

	exists2 := hashMap.Committed[*key]
	if exists2 != nil {
		return exists2.Data != nil
	}

	out := hashMap.Bucket.Get(key[:])
	hashMap.Committed[*key] = &CommitedMapElement{
		out,
		"view",
		"",
	}
	return out != nil
}

func (hashMap *HashMap) Update(key *[20]byte, data []byte) {

	exists := hashMap.Changes[*key]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[*key] = exists
	}
	exists.Data = data
	exists.Status = "update"

	return
}

func (hashMap *HashMap) Delete(key *[20]byte) {

	exists := hashMap.Changes[*key]
	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[*key] = exists
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
				committed = new(CommitedMapElement)
				hashMap.Committed[k] = committed
			}

			if v.Status == "del" && committed.Status != "del" {
				committed.Status = "del"
				committed.Commit = ""
				committed.Data = nil
			} else if v.Status == "update" && committed.Status != "update" {
				committed.Status = "update"
				committed.Commit = ""
				committed.Data = v.Data
			}

		}

	}
	hashMap.Changes = make(map[[20]byte]*ChangesMapElement)
}

func (hashMap *HashMap) Rollback() {
	hashMap.Changes = make(map[[20]byte]*ChangesMapElement)
}

func (hashMap *HashMap) WriteToStore() {

	for k, v := range hashMap.Committed {

		if v.Status == "del" {
			if err := hashMap.Bucket.Delete(k[:]); err != nil {
				panic(err)
			}
			v.Status = "view"
			v.Commit = "del"
		} else if v.Status == "update" {
			if err := hashMap.Bucket.Put(k[:], v.Data); err != nil {
				panic(err)
			}
			v.Commit = "update"
			v.Status = "view"
		}

	}

	return
}
