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
	Bucket    *bbolt.Bucket
	Virtual   map[string]*VirtualHashMapElement
	KeyLength int
}

func CreateNewHashMap(tx *bbolt.Tx, name string, keyLength int) (hashMap *HashMap, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	hashMap = &HashMap{
		Virtual:   make(map[string]*VirtualHashMapElement),
		Bucket:    tx.Bucket([]byte(name)),
		KeyLength: keyLength,
	}
	return
}

func (hashMap *HashMap) Get(key []byte) (out []byte) {

	keyStr := string(key)

	exists := hashMap.Virtual[keyStr]
	if exists != nil {
		out = exists.Data
		return
	}

	out = hashMap.Bucket.Get(key)
	hashMap.Virtual[keyStr] = &VirtualHashMapElement{
		out,
		"view",
		"",
	}
	return
}

func (hashMap *HashMap) Exists(key []byte) bool {
	keyStr := string(key)

	exists := hashMap.Virtual[keyStr]
	if exists != nil {
		return exists.Data != nil
	}
	out := hashMap.Bucket.Get(key)
	hashMap.Virtual[keyStr] = &VirtualHashMapElement{
		out,
		"view",
		"",
	}
	return out != nil
}

func (hashMap *HashMap) Update(key []byte, data []byte) {

	keyStr := string(key)

	exists := hashMap.Virtual[keyStr]
	if exists == nil {
		exists = new(VirtualHashMapElement)
		hashMap.Virtual[keyStr] = exists
	}
	exists.Data = data
	exists.Status = "update"

	return
}

func (hashMap *HashMap) Delete(key []byte) {

	keyStr := string(key)

	exists := hashMap.Virtual[keyStr]
	if exists == nil {
		exists = new(VirtualHashMapElement)
		hashMap.Virtual[keyStr] = exists
	}
	exists.Status = "del"
	exists.Data = nil
	return
}

func (hashMap *HashMap) Commit() (err error) {

	for k, v := range hashMap.Virtual {

		key := []byte(k)
		if len(key) != hashMap.KeyLength {
			return errors.New("KeyLength is invalid")
		}

		if v.Status == "del" {
			hashMap.Bucket.Delete(key)
			v.Status = "view"
			v.Committed = "del"
			v.Data = nil
		} else if v.Status == "update" {
			if err = hashMap.Bucket.Put(key, v.Data); err != nil {
				return
			}
			v.Committed = "update"
			v.Status = "view"
		}

	}

	return
}
