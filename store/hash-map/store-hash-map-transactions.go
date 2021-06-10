package hash_map

import (
	"encoding/json"
	"errors"
	"pandora-pay/helpers"
)

type transactionChange struct {
	Key        []byte
	Transition []byte
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) (err error) {

	values := make([]*transactionChange, 0)
	for k, v := range hashMap.Changes {
		if v.Status == "del" || v.Status == "update" {

			existsCommitted := hashMap.Committed[k]
			if existsCommitted != nil {
				values = append(values, &transactionChange{
					Key:        []byte(k),
					Transition: existsCommitted.Element.SerializeToBytes(),
				})
			} else {
				outData := hashMap.Tx.Get(hashMap.name + k)
				values = append(values, &transactionChange{
					Key:        []byte(k),
					Transition: outData,
				})
			}

		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		return
	}

	return hashMap.Tx.Put(hashMap.name+"transitions:"+prefix, marshal)
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) error {
	return hashMap.Tx.Delete(hashMap.name + "transitions:" + prefix)
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) (err error) {
	data := hashMap.Tx.Get(hashMap.name + "transitions:" + prefix)
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
