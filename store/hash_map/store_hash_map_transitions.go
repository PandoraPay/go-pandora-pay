package hash_map

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/helpers"
)

type transactionChange struct {
	Key        []byte
	Transition []byte
}

type transactionChanges struct {
	List []*transactionChange
}

func (hashMap *HashMap) WriteTransitionalChangesToStore(prefix string) (bool, error) {

	empty := true
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
					change.Transition = helpers.SerializeToBytes(existsCommitted.Element)
				}
			} else {
				//safe to Get because it will be cloned afterwards
				change.Transition = hashMap.Tx.Get(hashMap.name + ":map:" + k)
			}

			empty = false
			changes.List = append(changes.List, change)
		}
	}

	if empty {
		return false, nil
	}

	marshal, err := msgpack.Marshal(changes)
	if err != nil {
		return false, nil
	}

	hashMap.Tx.Put(hashMap.name+":transitions:"+prefix, marshal)

	return true, nil
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
	if err = msgpack.Unmarshal(data, changes); err != nil {
		return err
	}

	//in reverse
	for i := len(changes.List) - 1; i >= 0; i-- {

		change := changes.List[i]

		if change.Transition == nil {

			hashMap.Delete(string(change.Key))

		} else {

			var element HashMapElementSerializableInterface
			if element, err = hashMap.deserialize(change.Key, change.Transition, 0); err != nil {
				return
			}

			if err = hashMap.Update(string(change.Key), element); err != nil {
				return
			}
		}

	}

	return nil
}
