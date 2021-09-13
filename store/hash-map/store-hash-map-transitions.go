package hash_map

import (
	"encoding/binary"
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
				change.Transition = hashMap.Tx.Get(hashMap.name + k)
			}

			values = append(values, change)
		}
	}

	marshal, err := json.Marshal(values)
	if err != nil {
		return
	}

	if err = hashMap.Tx.Put(hashMap.name+":transitions:"+prefix, marshal); err != nil {
		return
	}

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, hashMap.Count)
	if err = hashMap.Tx.Put(hashMap.name+":transitionsCount:"+prefix, buf[:n]); err != nil {
		return
	}

	return nil
}

func (hashMap *HashMap) DeleteTransitionalChangesFromStore(prefix string) error {
	return hashMap.Tx.Delete(hashMap.name + ":transitions:" + prefix)
}

func (hashMap *HashMap) ReadTransitionalChangesFromStore(prefix string) (err error) {
	data := hashMap.Tx.Get(hashMap.name + ":transitions:" + prefix)
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
			if element, err = hashMap.Deserialize(change.Key, change.Transition); err != nil {
				return
			}

			hashMap.Committed[string(change.Key)] = &CommittedMapElement{
				Element: element,
				Status:  "update",
				Stored:  "",
			}

		}
	}

	data = hashMap.Tx.Get(hashMap.name + ":transitionsCount:" + prefix)
	if data == nil {
		return errors.New("transitionsCount didn't exist")
	}

	count, p := binary.Uvarint(data)
	if p <= 0 {
		return errors.New("Error reading")
	}
	hashMap.Count = count

	return nil
}
