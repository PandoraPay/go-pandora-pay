package store

import (
	"encoding/json"
	"pandora-pay/helpers"
)

type transactionChange struct {
	key        []byte
	transition []byte
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) {

	values := make([]transactionChange, 0)
	for key, v := range hashMap.Changes {
		if v.Status == "del" || v.Status == "update" {
			original := hashMap.get(&key, false)
			values = append(values, transactionChange{
				key:        key[:],
				transition: original,
			})

		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}

	hashMap.Bucket.Put([]byte("transitions_"+prefix), marshal)
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) {
	hashMap.Bucket.Delete([]byte("transitions_" + prefix))
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) {
	data := hashMap.Bucket.Get([]byte("transitions_" + prefix))
	if data == nil {
		panic("transitions didn't exist")
	}

	values := make([]transactionChange, 0)
	if err := json.Unmarshal(data, &values); err != nil {
		panic(err)
	}

	for _, change := range values {
		if change.transition == nil {
			hashMap.Committed[*helpers.Byte20(change.key)] = &CommitedMapElement{
				Data:   nil,
				Status: "del",
				Commit: "",
			}
		} else {
			hashMap.Committed[*helpers.Byte20(change.key)] = &CommitedMapElement{
				Data:   change.transition,
				Status: "update",
				Commit: "",
			}
		}
	}

}
