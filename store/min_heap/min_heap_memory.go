package min_heap

type MinHeapMemory struct {
	*MinHeap
}

func CreateMinHeapMemory() *MinHeapMemory {

	minHeap := CreateMinHeap()

	array := make([]*MinHeapElement, 0)
	size := uint64(0)

	minHeap.updateElement = func(index uint64, x *MinHeapElement) error {
		array[index] = x
		return nil
	}
	minHeap.addElement = func(x *MinHeapElement) {
		array = append(array, x)
		size += 1
	}
	minHeap.removeElement = func() {
		array = array[:size-1]
		size -= 1
	}
	minHeap.getElement = func(index uint64) (*MinHeapElement, error) {
		return array[index], nil
	}
	minHeap.getSize = func() uint64 {
		return size
	}

	return &MinHeapMemory{
		minHeap,
	}
}
