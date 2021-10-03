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
			change := &transactionChange{
				Key:        []byte(k),
				Transition: nil,
			}

			if existsCommitted != nil {
				if existsCommitted.Element != nil {
					change.Transition = existsCommitted.Element.SerializeToBytes()
				}
			} else {
				//safe to Get because it will be cloned afterwards
				change.Transition = hashMap.Tx.Get(hashMap.name + ":map:" + k)
			}

			values = append(values, change)
		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		return
	}

	if err = hashMap.Tx.PutClone(hashMap.name+":transitions:"+prefix, marshal); err != nil {
		return
	}

	return nil
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) (err error) {
	if err = hashMap.Tx.Delete(hashMap.name + ":transitions:" + prefix); err != nil {
		return
	}
	return hashMap.Tx.Delete(hashMap.name + ":transitionsCount:" + prefix)
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) (err error) {

	//Clone required to avoid changing the data afterwards
	data := hashMap.Tx.GetClone(hashMap.name + ":transitions:" + prefix)
	if data == nil {
		return errors.New("transitions didn't exist")
	}

	values := make([]*transactionChange, 0)
	if err = json.Unmarshal(data, &values); err != nil {
		return err
	}

	for _, change := range values {
		if change.Transition == nil {

			hashMap.Changes[string(change.Key)] = &ChangesMapElement{
				Element: nil,
				Status:  "del",
			}

		} else {

			var element helpers.SerializableInterface
			if element, err = hashMap.Deserialize(change.Key, change.Transition); err != nil {
				return
			}

			hashMap.Changes[string(change.Key)] = &ChangesMapElement{
				Element: element,
				Status:  "update",
			}

		}
	}

	return nil
}
