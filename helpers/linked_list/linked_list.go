package linked_list

type linkedListItem struct {
	Next *linkedListItem
	Data interface{}
}

type LinkedList struct {
	First  *linkedListItem
	Last   *linkedListItem
	Length int
}

func (list *LinkedList) Empty() {
	list.First = nil
	list.Last = nil
	list.Length = 0
}

func (list *LinkedList) PushFront(data interface{}) {
	if list.First == nil {
		list.First = &linkedListItem{nil, data}
		list.Last = list.First
	} else {
		next := &linkedListItem{list.First, data}
		list.First = next
	}
	list.Length++
}

func (list *LinkedList) Push(data interface{}) {
	if list.First == nil {
		list.First = &linkedListItem{nil, data}
		list.Last = list.First
	} else {
		next := &linkedListItem{nil, data}
		list.Last.Next = next
		list.Last = next
	}
	list.Length++
}

func (list *LinkedList) PopFirst() interface{} {
	if list.First != nil {
		data := list.First.Data
		list.First = list.First.Next
		list.Length--
		return data
	}
	return nil
}

func (list *LinkedList) GetFirst() interface{} {
	if list.First != nil {
		return list.First.Data
	}
	return nil
}

func (list *LinkedList) GetLast() interface{} {
	if list.Last != nil {
		return list.Last.Data
	}
	return nil
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}
