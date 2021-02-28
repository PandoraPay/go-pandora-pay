package store

import (
	"errors"
	"go.etcd.io/bbolt"
)

type VirtualHashMapElement struct {
	Data      []byte
	Status    string
	Committed string
}

type HashMap struct {
	Bucket  *bbolt.Bucket
	Virtual map[string]*VirtualHashMapElement
}

func CreateNewHashMap(tx *bbolt.Tx, name string) (hashMap *HashMap, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	hashMap = new(HashMap)
	hashMap.Virtual = make(map[string]*VirtualHashMapElement)
	hashMap.Bucket = tx.Bucket([]byte(name))
	return

}

func (hashMap *HashMap) Get(key string) (out []byte) {

	exists := hashMap.Virtual[key]
	if exists != nil {
		out = exists.Data
		return
	}

	out = hashMap.Bucket.Get([]byte(key))
	hashMap.Virtual[key] = &VirtualHashMapElement{
		out,
		"view",
		"",
	}
	return
}

func (hashMap *HashMap) Exists(key string) bool {
	exists := hashMap.Virtual[key]
	if exists != nil {
		return exists.Data != nil
	}
	out := hashMap.Bucket.Get([]byte(key))
	hashMap.Virtual[key] = &VirtualHashMapElement{
		out,
		"view",
		"",
	}
	return out != nil
}

func (hashMap *HashMap) Update(key string, data []byte) {

	exists := hashMap.Virtual[key]
	if exists == nil {
		exists = new(VirtualHashMapElement)
		hashMap.Virtual[key] = exists
	}
	exists.Data = data
	exists.Status = "update"

	return
}

func (hashMap *HashMap) Delete(key string) {

	exists := hashMap.Virtual[key]
	if exists == nil {
		exists = new(VirtualHashMapElement)
		hashMap.Virtual[key] = exists
	}
	exists.Status = "del"
	exists.Data = nil
	return
}

func (hashMap *HashMap) Commit() (err error) {

	for k, v := range hashMap.Virtual {

		if v.Status == "del" {
			hashMap.Bucket.Delete([]byte(k))
			v.Status = "view"
			v.Committed = "del"
			v.Data = nil
		} else if v.Status == "update" {
			if err = hashMap.Bucket.Put([]byte(k), v.Data); err != nil {
				return
			}
			v.Committed = "update"
			v.Status = "view"
		}

	}

	return
}
