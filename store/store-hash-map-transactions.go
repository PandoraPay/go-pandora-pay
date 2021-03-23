package store

import (
	"encoding/json"
	"errors"
)

type transactionChange struct {
	key        []byte
	transition []byte
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) error {

	values := []transactionChange{}
	for k, v := range hashMap.Changes {
		if v.Status == "del" || v.Status == "update" {
			key := []byte(k)
			original := hashMap.get(key, false)
			values = append(values, transactionChange{
				key:        key,
				transition: original,
			})

		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return hashMap.Bucket.Put([]byte("transitions_"+prefix), marshal)
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) error {
	return hashMap.Bucket.Delete([]byte("transitions_" + prefix))
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) error {
	data := hashMap.Bucket.Get([]byte("transitions_" + prefix))
	if data == nil {
		return errors.New("transitions didn't exist")
	}

	values := []transactionChange{}
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	for _, change := range values {
		if change.transition == nil {
			hashMap.Committed[string(change.key)] = &CommittedMapElement{
				Data:   nil,
				Status: "del",
				Commit: "",
			}
		} else {
			hashMap.Committed[string(change.key)] = &CommittedMapElement{
				Data:   change.transition,
				Status: "update",
				Commit: "",
			}
		}
	}

	return nil
}
