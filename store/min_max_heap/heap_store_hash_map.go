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

	if err := self.Delete(found.(*HeapDictElement).Index); err != nil {
		return err
	}

	self.DictMap.Delete(string(key))
	return nil
}

func (self *HeapStoreHashMap) GetKey(Key []byte) (*HeapDictElement, error) {
	found, err := self.DictMap.Get(string(Key))
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, errors.New("Key is not found")
	}

	return found.(*HeapDictElement), nil
}

func NewHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string, compare func(a, b float64) bool) *HeapStoreHashMap {

	heap := NewHeap(compare)
	hashMap := hash_map.CreateNewHashMap(dbTx, name, 0, false)
	dictMap := hash_map.CreateNewHashMap(dbTx, name+"_dict", 0, false)

	hashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return &HeapElement{HeapKey: key}, nil
	}

	dictMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return &HeapDictElement{Key: key}, nil
	}

	heap.updateElement = func(index uint64, x *HeapElement) (err error) {
		if err = hashMap.Update(strconv.FormatUint(index, 10), x); err != nil {
			return
		}
		return dictMap.Update(string(x.Key), &HeapDictElement{nil, x.Key, index})
	}

	heap.addElement = func(x *HeapElement) (err error) {
		index := hashMap.Count
		if err = hashMap.Update(strconv.FormatUint(index, 10), x); err != nil {
			return
		}
		return dictMap.Update(string(x.Key), &HeapDictElement{nil, x.Key, index})
	}

	heap.removeElement = func() (*HeapElement, error) {

		if hashMap.Count == 0 {
			return nil, nil
		}

		index := hashMap.Count - 1

		x, err := heap.getElement(index)
		if err != nil {
			return x, err
		}

		hashMap.Delete(strconv.FormatUint(index, 10))
		dictMap.Delete(string(x.Key))

		return x, nil
	}

	heap.getElement = func(index uint64) (*HeapElement, error) {
		el, err := hashMap.Get(strconv.FormatUint(index, 10))
		if err != nil || el == nil {
			return nil, err
		}
		return el.(*HeapElement), nil
	}

	heap.GetSize = func() uint64 {
		return hashMap.Count
	}

	return &HeapStoreHashMap{
		heap,
		hashMap,
		dictMap,
	}
}

func NewMinHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string) *HeapStoreHashMap {
	return NewHeapStoreHashMap(dbTx, name, func(a, b float64) bool {
		return a < b
	})
}

func NewMaxHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string) *HeapStoreHashMap {
	return NewHeapStoreHashMap(dbTx, name, func(a, b float64) bool {
		return b < a
	})
}
