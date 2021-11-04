package min_heap

import "pandora-pay/helpers"

// based on https://golangbyexample.com/minheap-in-golang/

type MinHeap struct {
	getElement    func(index uint64) (*MinHeapElement, error)
	updateElement func(index uint64, x *MinHeapElement) error
	addElement    func(x *MinHeapElement)
	removeElement func()
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

func (m *MinHeap) Insert(score uint64, data helpers.SerializableInterface) error {
	m.addElement(&MinHeapElement{nil, data, score})
	return m.upHeapify(m.getSize() - 1)
}

func (m *MinHeap) swap(first, second uint64) error {
	temp, err := m.getElement(first)
	if err != nil {
		return err
	}

	secondEl, err := m.getElement(second)
	if err != nil {
		return err
	}
	if err := m.updateElement(first, secondEl); err != nil {
		return err
	}
	return m.updateElement(second, temp)
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

	if a, err = m.getElement(leftChildIndex); err != nil {
		return
	}
	if b, err = m.getElement(smallest); err != nil {
		return
	}
	if leftChildIndex < m.getSize() && a.Score < b.Score {
		smallest = leftChildIndex
	}

	if a, err = m.getElement(rightRightIndex); err != nil {
		return
	}
	if b, err = m.getElement(smallest); err != nil {
		return
	}
	if rightRightIndex < m.getSize() && a.Score < b.Score {
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

func (m *MinHeap) RemoveMin() (*MinHeapElement, error) {
	top, err := m.getElement(0)
	if err != nil {
		return nil, err
	}

	last, err := m.getElement(m.getSize() - 1)
	if err != nil {
		return nil, err
	}

	if err = m.updateElement(0, last); err != nil {
		return nil, err
	}
	m.removeElement()
	if err := m.downHeapify(0); err != nil {
		return nil, err
	}
	return top, nil
}

func (m *MinHeap) GetMin() (*MinHeapElement, error) {
	return m.getElement(0)
}

func CreateMinHeap() *MinHeap {
	return &MinHeap{}
}
