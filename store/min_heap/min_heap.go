package min_heap

// based on https://golangbyexample.com/minheap-in-golang/

type MinHeap struct {
	getElement    func(index uint64) (*MinHeapElement, error)
	updateElement func(index uint64, x *MinHeapElement)
	addElement    func(x *MinHeapElement)
	removeElement func() (*MinHeapElement, error)
	getSize       func() uint64
}

func (m *MinHeap) leaf(index uint64) bool {
	return index >= (m.getSize()/2) && index <= m.getSize()
}

func (m *MinHeap) parent(index uint64) uint64 {
	if index == 0 {
		return 0
	}
	return (index - 1) / 2
}

func (m *MinHeap) leftchild(index uint64) uint64 {
	return 2*index + 1
}

func (m *MinHeap) rightchild(index uint64) uint64 {
	return 2*index + 2
}

func (m *MinHeap) Insert(score uint64, key []byte) error {
	m.addElement(&MinHeapElement{nil, key, score})
	return m.upHeapify(m.getSize() - 1)
}

func (m *MinHeap) swap(first, second uint64) error {
	firstEl, err := m.getElement(first)
	if err != nil {
		return err
	}

	secondEl, err := m.getElement(second)
	if err != nil {
		return err
	}
	m.updateElement(first, secondEl)
	m.updateElement(second, firstEl)
	return nil
}

func (m *MinHeap) upHeapify(index uint64) (err error) {
	var a, b *MinHeapElement
	for {
		if a, err = m.getElement(index); err != nil {
			return
		}
		if b, err = m.getElement(m.parent(index)); err != nil {
			return
		}

		if a.Score >= b.Score {
			return
		}
		if err = m.swap(index, m.parent(index)); err != nil {
			return
		}
		index = m.parent(index)
	}
}

func (m *MinHeap) downHeapify(current uint64) (err error) {
	if m.leaf(current) {
		return
	}

	var a, b *MinHeapElement

	smallest := current
	leftChildIndex := m.leftchild(current)
	rightRightIndex := m.rightchild(current)
	//If current is smallest then return

	if leftChildIndex < m.getSize() {
		if a, err = m.getElement(leftChildIndex); err != nil {
			return
		}
		if b, err = m.getElement(smallest); err != nil {
			return
		}
		if a.Score < b.Score {
			smallest = leftChildIndex
		}
	}

	if rightRightIndex < m.getSize() {
		if a.Score < b.Score {
			if a, err = m.getElement(rightRightIndex); err != nil {
				return
			}
			if b, err = m.getElement(smallest); err != nil {
				return
			}
		}
		smallest = rightRightIndex
	}
	if smallest != current {
		if err = m.swap(current, smallest); err != nil {
			return
		}
		if err = m.downHeapify(smallest); err != nil {
			return
		}
	}
	return
}

//https://stackoverflow.com/a/12664523/14319261
func (m *MinHeap) Delete(index uint64) error {

	element, err := m.removeElement()
	if err != nil {
		return err
	}

	if index == m.getSize() {
		return nil
	}

	m.updateElement(index, element)

	if index > 0 {
		middle, err := m.getElement((index - 1) / 2)
		if err != nil {
			return err
		}

		if index > 0 && element.Score > middle.Score {
			return m.upHeapify(index)
		}

	}

	if index < m.getSize()/2 {
		return m.downHeapify(index)
	}

	return nil
}

func (m *MinHeap) RemoveMin() (*MinHeapElement, error) {

	if err := m.swap(0, m.getSize()-1); err != nil {
		return nil, err
	}

	top, err := m.removeElement()
	if err != nil {
		return nil, err
	}

	if err := m.downHeapify(0); err != nil {
		return nil, err
	}
	return top, nil
}

func (m *MinHeap) GetMin() (*MinHeapElement, error) {
	return m.getElement(0)
}

func NewMinHeap() *MinHeap {
	return &MinHeap{}
}
