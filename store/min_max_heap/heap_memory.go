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

	if err := self.Delete(value); err != nil {
		return err
	}

	delete(self.dict, string(key))
	return nil
}

func (self *HeapMemory) Update(score float64, key []byte) error {
	value, ok := self.dict[string(key)]
	if ok {
		return self.Delete(value)
	}
	return self.Insert(score, key)
}

func NewHeapMemory(compare func(a, b float64) bool) *HeapMemory {

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
		if index >= uint64(len(array)) {
			return nil, nil
		}
		return array[index], nil
	}
	heap.GetSize = func() uint64 {
		return size
	}

	return &HeapMemory{
		heap,
		dict,
	}
}

func NewMinMemoryHeap(name string) *HeapMemory {
	return NewHeapMemory(func(a, b float64) bool {
		return a < b
	})
}

func NewMaxMemoryHeap() *HeapMemory {
	return NewHeapMemory(func(a, b float64) bool {
		return b < a
	})
}
