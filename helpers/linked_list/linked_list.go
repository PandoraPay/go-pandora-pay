package linked_list

import "pandora-pay/helpers"

type linkedListItem[T any] struct {
	Next *linkedListItem[T]
	Data T
}

type LinkedList[T any] struct {
	First  *linkedListItem[T]
	Last   *linkedListItem[T]
	Length int
}

func (list *LinkedList[T]) Empty() {
	list.First = nil
	list.Last = nil
	list.Length = 0
}

func (list *LinkedList[T]) PushFront(data T) {
	if list.First == nil {
		list.First = &linkedListItem[T]{nil, data}
		list.Last = list.First
	} else {
		next := &linkedListItem[T]{list.First, data}
		list.First = next
	}
	list.Length++
}

func (list *LinkedList[T]) Push(data T) {
	if list.First == nil {
		list.First = &linkedListItem[T]{nil, data}
		list.Last = list.First
	} else {
		next := &linkedListItem[T]{nil, data}
		list.Last.Next = next
		list.Last = next
	}
	list.Length++
}

func (list *LinkedList[T]) PopFirst() (T, bool) {
	if list.First != nil {
		data := list.First.Data
		list.First = list.First.Next
		list.Length--
		return data, true
	}
	return helpers.Zero[T](), false
}

func (list *LinkedList[T]) GetFirst() (T, bool) {
	if list.First != nil {
		return list.First.Data, true
	}
	return helpers.Zero[T](), false
}

func (list *LinkedList[T]) GetLast() (T, bool) {
	if list.Last != nil {
		return list.Last.Data, true
	}
	return helpers.Zero[T](), false
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}
