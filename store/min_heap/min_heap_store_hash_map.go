package min_heap

import (
	"errors"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type MinHeapStoreHashMap struct {
	*MinHeap
	hashMap *hash_map.HashMap
	dictMap *hash_map.HashMap
}

func (self *MinHeapStoreHashMap) DeleteByKey(key []byte) error {
	found, err := self.dictMap.Get(string(key))
	if err != nil {
		return err
	}
	if found == nil {
		return errors.New("Key is not found")
	}

	return self.Delete(found.(*MinHeapDictElement).Index)
}

func NewMinHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string) *MinHeapStoreHashMap {

	minHeap := NewMinHeap()
	hashMap := hash_map.CreateNewHashMap(dbTx, name, 0, false)
	dictMap := hash_map.CreateNewHashMap(dbTx, name+"_dict", 0, false)

	newSize := hashMap.Count

	minHeap.updateElement = func(index uint64, x *MinHeapElement) {
		hashMap.Update(strconv.FormatUint(index, 10), x)
		dictMap.Update(string(x.Key), &MinHeapDictElement{nil, index})
		return
	}

	minHeap.addElement = func(x *MinHeapElement) {
		hashMap.Update(strconv.FormatUint(newSize, 10), x)
		dictMap.Update(string(x.Key), &MinHeapDictElement{nil, newSize})
		newSize += 1
		return
	}

	minHeap.removeElement = func() (*MinHeapElement, error) {
		newSize -= 1

		x, err := minHeap.getElement(newSize)
		if err != nil {
			return x, err
		}

		hashMap.Delete(strconv.FormatUint(newSize, 10))
		dictMap.Delete(string(x.Key))

		return x, nil
	}

	minHeap.getElement = func(index uint64) (*MinHeapElement, error) {
		el, err := hashMap.Get(strconv.FormatUint(index, 10))
		if err != nil {
			return nil, err
		}
		return el.(*MinHeapElement), nil
	}

	minHeap.getSize = func() uint64 {
		return newSize
	}

	return &MinHeapStoreHashMap{
		minHeap,
		hashMap,
		dictMap,
	}
}
