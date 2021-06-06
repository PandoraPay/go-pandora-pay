package hash_map

import (
	"encoding/json"
	"errors"
	"pandora-pay/helpers"
)

type transactionChange struct {
	Key        string
	Transition []byte
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) (err error) {

	values := make([]*transactionChange, 0)
	for k, v := range hashMap.Changes {
		if v.Status == "del" || v.Status == "update" {

			existsCommitted := hashMap.Committed[k]
			if existsCommitted != nil {
				values = append(values, &transactionChange{
					Key:        k,
					Transition: existsCommitted.Element.SerializeToBytes(),
				})
			} else {
				outData := hashMap.Tx.Get(k)
				values = append(values, &transactionChange{
					Key:        k,
					Transition: outData,
				})
			}

		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		return
	}

	return hashMap.Tx.Put("transitions_"+prefix, marshal)
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) error {
	return hashMap.Tx.Delete("transitions_" + prefix)
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) (err error) {
	data := hashMap.Tx.Get("transitions_" + prefix)
	if data == nil {
		return errors.New("transitions didn't exist")
	}

	values := make([]*transactionChange, 0)
	if err = json.Unmarshal(data, &values); err != nil {
		return err
	}

	for _, change := range values {
		if change.Transition == nil {
			hashMap.Committed[string(change.Key)] = &CommittedMapElement{
				Element: nil,
				Status:  "del",
				Stored:  "",
			}
		} else {
			var element helpers.SerializableInterface
			if element, err = hashMap.Deserialize(change.Transition); err != nil {
				return
			}

			hashMap.Committed[string(change.Key)] = &CommittedMapElement{
				Element: element,
				Status:  "update",
				Stored:  "",
			}
		}
	}

	return nil
}
