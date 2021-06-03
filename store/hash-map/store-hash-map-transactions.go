package hash_map

import (
	"encoding/json"
	"errors"
)

type transactionChange struct {
	Key        string
	Transition []byte
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) error {

	values := make([]*transactionChange, 0)
	for k, v := range hashMap.Changes {
		if v.Status == "del" || v.Status == "update" {
			original := hashMap.get(k, false)
			values = append(values, &transactionChange{
				Key:        k,
				Transition: original,
			})

		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return hashMap.Tx.Put("transitions_"+prefix, marshal)
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) error {
	return hashMap.Tx.Delete("transitions_" + prefix)
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) error {
	data := hashMap.Tx.Get("transitions_" + prefix)
	if data == nil {
		return errors.New("transitions didn't exist")
	}

	values := make([]*transactionChange, 0)
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	for _, change := range values {
		if change.Transition == nil {
			hashMap.Committed[string(change.Key)] = &CommittedMapElement{
				Data:   nil,
				Status: "del",
				Stored: "",
			}
		} else {
			hashMap.Committed[string(change.Key)] = &CommittedMapElement{
				Data:   change.Transition,
				Status: "update",
				Stored: "",
			}
		}
	}

	return nil
}
