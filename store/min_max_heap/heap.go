package min_max_heap

// based on https://golangbyexample.com/Heap-in-golang/

type Heap struct {
	getElement    func(index uint64) (*HeapElement, error)
	updateElement func(index uint64, x *HeapElement) error
	addElement    func(x *HeapElement) error
	removeElement func() (*HeapElement, error)
	GetSize       func() uint64
	compare       func(a, b float64) bool
}

func (m *Heap) leaf(index uint64) bool {
	return index >= (m.GetSize()/2) && index <= m.GetSize()
}

func (m *Heap) parent(index uint64) uint64 {
	if index == 0 {
		return 0
	}
	return (index - 1) / 2
}

func (m *Heap) leftchild(index uint64) uint64 {
	return 2*index + 1
}

func (m *Heap) rightchild(index uint64) uint64 {
	return 2*index + 2
}

func (m *Heap) Insert(score float64, key []byte) error {
	if err := m.addElement(&HeapElement{nil, nil, key, score}); err != nil {
		return err
	}
	return m.upHeapify(m.GetSize() - 1)
}

func (m *Heap) swap(first, second uint64) error {
	firstEl, err := m.getElement(first)
	if err != nil {
		return err
	}

	secondEl, err := m.getElement(second)
	if err != nil {
		return err
	}

	if err = m.updateElement(first, secondEl); err != nil {
		return err
	}
	return m.updateElement(second, firstEl)
}

func (m *Heap) upHeapify(index uint64) (err error) {
	var a, b *HeapElement
	for {
		if a, err = m.getElement(index); err != nil {
			return
		}
		if b, err = m.getElement(m.parent(index)); err != nil {
			return
		}

		if !m.compare(a.Score, b.Score) {
			return
		}
		if err = m.swap(index, m.parent(index)); err != nil {
			return
		}
		index = m.parent(index)
	}
}

func (m *Heap) downHeapify(current uint64) (err error) {
	if m.leaf(current) {
		return
	}

	var a, b *HeapElement

	smallest := current
	leftChildIndex := m.leftchild(current)
	rightRightIndex := m.rightchild(current)
	//If current is smallest then return

	if leftChildIndex < m.GetSize() {
		if a, err = m.getElement(leftChildIndex); err != nil {
			return
		}
		if b, err = m.getElement(smallest); err != nil {
			return
		}
		if m.compare(a.Score, b.Score) {
			smallest = leftChildIndex
		}
	}

	if rightRightIndex < m.GetSize() {
		if a, err = m.getElement(rightRightIndex); err != nil {
			return
		}
		if b, err = m.getElement(smallest); err != nil {
			return
		}
		if m.compare(a.Score, b.Score) {
			smallest = rightRightIndex
		}
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

// https://stackoverflow.com/a/12664523/14319261
func (m *Heap) Delete(index uint64) error {

	if m.GetSize() == 0 {
		return nil
	}

	if m.GetSize() == 1 {
		_, err := m.removeElement()
		return err
	}

	element, err := m.removeElement()
	if err != nil {
		return err
	}

	if err = m.updateElement(index, element); err != nil {
		return err
	}

	if index > 1 {
		p, err := m.getElement(m.parent(index))
		if err != nil {
			return err
		}

		if m.compare(element.Score, p.Score) {
			return m.upHeapify(index)
		} else {
			return m.downHeapify(index)
		}
	}

	return m.downHeapify(0)
}

func (m *Heap) RemoveTop() (*HeapElement, error) {

	if err := m.swap(0, m.GetSize()-1); err != nil {
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

func (m *Heap) GetTop() (*HeapElement, error) {
	return m.getElement(0)
}

/*
Minheap

	func (a,b uint64) bool{
		return a < b
	}

Maxheap

	func (a,b uint64) bool{
		return b < a
	}
*/
func NewHeap(compare func(a, b float64) bool) *Heap {
	return &Heap{
		compare: compare,
	}
}
