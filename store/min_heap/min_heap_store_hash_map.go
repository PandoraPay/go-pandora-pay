package min_heap

import (
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type MinHeapStoreHashMap struct {
	*MinHeap
	hashMap *hash_map.HashMap
}

func CreateMinHeapStoreHashMap(dbTx store_db_interface.StoreDBTransactionInterface, name string) *MinHeapStoreHashMap {

	minHeap := CreateMinHeap()
	hashMap := hash_map.CreateNewHashMap(dbTx, name, 0, false)

	newSize := hashMap.Count

	minHeap.updateElement = func(index uint64, x *MinHeapElement) error {
		hashMap.Update(strconv.FormatUint(index, 10), x)
		return nil
	}

	minHeap.addElement = func(x *MinHeapElement) {
		hashMap.Update(strconv.FormatUint(newSize, 10), x)
		newSize += 1
		return
	}

	minHeap.removeElement = func() {
		hashMap.Delete(strconv.FormatUint(newSize, 10))
		newSize -= 1
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
	}
}
