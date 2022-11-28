package min_max_heap

import "errors"

type HeapMemory struct {
	*Heap
	size  uint64
	array []*HeapElement
	dict  map[string]uint64
}

func (self *HeapMemory) Reset() {
	self.array = make([]*HeapElement, 0)
	self.dict = make(map[string]uint64)
	self.size = 0
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
		if err := self.Delete(value); err != nil {
			return err
		}
	}
	return self.Insert(score, key)
}

func NewHeapMemory(compare func(a, b float64) bool) *HeapMemory {

	heap := &HeapMemory{
		NewHeap(compare),
		0,
		make([]*HeapElement, 0),
		make(map[string]uint64),
	}

	heap.updateElement = func(index uint64, x *HeapElement) error {
		if index < uint64(len(heap.array)) {
			heap.array[index] = x
		} else {
			heap.array = append(heap.array, x)
		}
		heap.dict[string(x.Key)] = index
		return nil
	}
	heap.addElement = func(x *HeapElement) error {
		heap.array = append(heap.array, x)
		heap.dict[string(x.Key)] = heap.size
		heap.size += 1
		return nil
	}
	heap.removeElement = func() (*HeapElement, error) {

		if heap.size == 0 {
			return nil, nil
		}

		heap.size -= 1

		x := heap.array[heap.size]
		heap.array = heap.array[:heap.size]
		delete(heap.dict, string(x.Key))

		return x, nil
	}
	heap.getElement = func(index uint64) (*HeapElement, error) {
		if index >= uint64(len(heap.array)) {
			return nil, nil
		}
		return heap.array[index], nil
	}
	heap.GetSize = func() uint64 {
		return heap.size
	}

	return heap
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
