package min_max_heap

import (
	"errors"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type HeapStoreHashMap struct {
	*Heap
	HashMap *hash_map.HashMap
	DictMap *hash_map.HashMap
}

func (self *HeapStoreHashMap) DeleteByKey(key []byte) error {
	found, err := self.DictMap.Get(string(key))
	if err != nil {
		return err
	}
	if found == nil {
		return errors.New("Key is not found")
	}

	return self.Delete(found.(*HeapDictElement).Index)
}

func NewHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string, compare func(a, b uint64) bool) *HeapStoreHashMap {

	heap := NewHeap(compare)
	hashMap := hash_map.CreateNewHashMap(dbTx, name, 0, false)
	dictMap := hash_map.CreateNewHashMap(dbTx, name+"_dict", 0, false)

	newSize := hashMap.Count

	heap.updateElement = func(index uint64, x *HeapElement) (err error) {
		if err = hashMap.Update(strconv.FormatUint(index, 10), x); err != nil {
			return
		}
		return dictMap.Update(string(x.Key), &HeapDictElement{nil, index})
	}

	heap.addElement = func(x *HeapElement) (err error) {
		if err = hashMap.Update(strconv.FormatUint(newSize, 10), x); err != nil {
			return
		}
		if err = dictMap.Update(string(x.Key), &HeapDictElement{nil, newSize}); err != nil {
			return
		}
		newSize += 1
		return nil
	}

	heap.removeElement = func() (*HeapElement, error) {
		newSize -= 1

		x, err := heap.getElement(newSize)
		if err != nil {
			return x, err
		}

		hashMap.Delete(strconv.FormatUint(newSize, 10))
		dictMap.Delete(string(x.Key))

		return x, nil
	}

	heap.getElement = func(index uint64) (*HeapElement, error) {
		el, err := hashMap.Get(strconv.FormatUint(index, 10))
		if err != nil {
			return nil, err
		}
		return el.(*HeapElement), nil
	}

	heap.getSize = func() uint64 {
		return newSize
	}

	return &HeapStoreHashMap{
		heap,
		hashMap,
		dictMap,
	}
}

func NewMinHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string) *HeapStoreHashMap {
	return NewHeapStoreHashMap(dbTx, name, func(a, b uint64) bool {
		return a < b
	})
}

func NewMaxHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string) *HeapStoreHashMap {
	return NewHeapStoreHashMap(dbTx, name, func(a, b uint64) bool {
		return b < a
	})
}
