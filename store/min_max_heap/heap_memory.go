package min_max_heap

import "errors"

type HeapMemory struct {
	*Heap
	dict map[string]uint64
}

func (self *HeapMemory) DeleteByKey(key []byte) error {
	value, ok := self.dict[string(key)]
	if !ok {
		return errors.New("Key is not found")
	}

	return self.Delete(value)
}

func NewHeapMemory(compare func(a, b uint64) bool) *HeapMemory {

	heap := NewHeap(compare)

	array := make([]*HeapElement, 0)
	dict := make(map[string]uint64)

	size := uint64(0)

	heap.updateElement = func(index uint64, x *HeapElement) error {
		array[index] = x
		dict[string(x.Key)] = index
		return nil
	}
	heap.addElement = func(x *HeapElement) error {
		array = append(array, x)
		dict[string(x.Key)] = size
		size += 1
		return nil
	}
	heap.removeElement = func() (*HeapElement, error) {
		size -= 1

		x := array[size]
		array = array[:size]
		delete(dict, string(x.Key))

		return x, nil
	}
	heap.getElement = func(index uint64) (*HeapElement, error) {
		return array[index], nil
	}
	heap.getSize = func() uint64 {
		return size
	}

	return &HeapMemory{
		heap,
		dict,
	}
}
