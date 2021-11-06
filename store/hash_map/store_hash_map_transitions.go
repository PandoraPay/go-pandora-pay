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

type transactionChanges struct {
	List []*transactionChange
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) (err error) {

	changes := &transactionChanges{}
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

			changes.List = append(changes.List, change)
		}
	}

	marshal, err := json.Marshal(changes)
	if err != nil {
		return
	}

	hashMap.Tx.PutClone(hashMap.name+":transitions:"+prefix, marshal)

	return nil
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) {
	hashMap.Tx.Delete(hashMap.name + ":transitions:" + prefix)
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) (err error) {

	//Clone required to avoid changing the data afterwards
	data := hashMap.Tx.Get(hashMap.name + ":transitions:" + prefix)
	if data == nil {
		return errors.New("transitions didn't exist")
	}

	changes := &transactionChanges{}
	if err = json.Unmarshal(data, &changes); err != nil {
		return err
	}

	//in reverse
	for i := len(changes.List) - 1; i >= 0; i-- {

		change := changes.List[i]

		if change.Transition == nil {

			hashMap.Delete(string(change.Key))

		} else {

			var element helpers.SerializableInterface
			if element, err = hashMap.Deserialize(change.Key, change.Transition); err != nil {
				return
			}

			if err = hashMap.Update(string(change.Key), element); err != nil {
				return
			}
		}

	}

	return nil
}
