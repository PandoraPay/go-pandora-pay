package min_heap

import "errors"

type MinHeapMemory struct {
	*MinHeap
	dict map[string]uint64
}

func (self *MinHeapMemory) DeleteByKey(key []byte) error {
	value, ok := self.dict[string(key)]
	if !ok {
		return errors.New("Key is not found")
	}

	return self.Delete(value)
}

func NewMinHeapMemory() *MinHeapMemory {

	minHeap := NewMinHeap()

	array := make([]*MinHeapElement, 0)
	dict := make(map[string]uint64)

	size := uint64(0)

	minHeap.updateElement = func(index uint64, x *MinHeapElement) {
		array[index] = x
		dict[string(x.Key)] = index
	}
	minHeap.addElement = func(x *MinHeapElement) {
		array = append(array, x)
		dict[string(x.Key)] = size
		size += 1
	}
	minHeap.removeElement = func() (*MinHeapElement, error) {
		size -= 1

		x := array[size]
		array = array[:size]
		delete(dict, string(x.Key))

		return x, nil
	}
	minHeap.getElement = func(index uint64) (*MinHeapElement, error) {
		return array[index], nil
	}
	minHeap.getSize = func() uint64 {
		return size
	}

	return &MinHeapMemory{
		minHeap,
		dict,
	}
}
