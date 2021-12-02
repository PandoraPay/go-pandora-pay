package linked_list

type linkedListItem struct {
	next *linkedListItem
	data interface{}
}

type LinkedList struct {
	First *linkedListItem
	last  *linkedListItem
}

func (list *LinkedList) PushFront(data interface{}) {
	if list.First == nil {
		list.First = &linkedListItem{nil, data}
		list.last = list.First
	} else {
		next := &linkedListItem{list.First, data}
		list.First = next
	}
}

func (list *LinkedList) Push(data interface{}) {
	if list.First == nil {
		list.First = &linkedListItem{nil, data}
		list.last = list.First
	} else {
		next := &linkedListItem{nil, data}
		list.last.next = next
		list.last = next
	}
}

func (list *LinkedList) PopFirst() interface{} {
	if list.First != nil {
		data := list.First.data
		list.First = list.First.next
		return data
	}
	return nil
}

func (list *LinkedList) GetFirst() interface{} {
	if list.First != nil {
		return list.First.data
	}
	return nil
}

func (list *LinkedList) GetLast() interface{} {
	if list.last != nil {
		return list.last.data
	}
	return nil
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}
