package store

import "encoding/json"

type transactionChange struct {
	key        []byte
	transition []byte
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) {

	values := make([]transactionChange, 0)
	for key, v := range hashMap.Changes {
		if v.Status == "del" || v.Status == "update" {
			original := hashMap.get([]byte(key), false)
			values = append(values, transactionChange{
				key:        []byte(key),
				transition: original,
			})

		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}

	if err := hashMap.Bucket.Put([]byte("transitions_"+prefix), marshal); err != nil {
		panic(err)
	}
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) {
	data := hashMap.Bucket.Get([]byte("transitions_" + prefix))
	if data == nil {
		panic("transitions didn't exist")
	}

	values := make([]transactionChange, 0)
	if err := json.Unmarshal(data, values); err != nil {
		panic(err)
	}

	for _, change := range values {
		if change.transition == nil {
			hashMap.Committed[string(change.key)] = &CommitedMapElement{
				Data:   nil,
				Status: "del",
				Commit: "",
			}
		} else {
			hashMap.Committed[string(change.key)] = &CommitedMapElement{
				Data:   change.transition,
				Status: "update",
				Commit: "",
			}
		}
	}

}
