package linked_list

import (
	"pandora-pay/helpers/generics"
)

type linkedListItem[T any] struct {
	Next *linkedListItem[T] `json:"next" msgpack:"next"`
	Data T                  `json:"data" msgpack:"data"`
}

type LinkedList[T any] struct {
	Head   *linkedListItem[T] `json:"head" msgpack:"head"`
	Tail   *linkedListItem[T] `json:"tail" msgpack:"tail"`
	Length int                `json:"length" msgpack:"length"`
}

func (list *LinkedList[T]) Empty() {
	list.Head = nil
	list.Tail = nil
	list.Length = 0
}

func (list *LinkedList[T]) PushFront(data T) {
	if list.Head == nil {
		list.Head = &linkedListItem[T]{nil, data}
		list.Tail = list.Head
	} else {
		next := &linkedListItem[T]{list.Head, data}
		list.Head = next
	}
	list.Length++
}

func (list *LinkedList[T]) Push(data T) {
	if list.Head == nil {
		list.Head = &linkedListItem[T]{nil, data}
		list.Tail = list.Head
	} else {
		next := &linkedListItem[T]{nil, data}
		list.Tail.Next = next
		list.Tail = next
	}
	list.Length++
}

func (list *LinkedList[T]) PopHead() (T, bool) {
	if list.Head != nil {
		data := list.Head.Data
		list.Head = list.Head.Next
		list.Length--
		return data, true
	}
	return generics.Zero[T](), false
}

func (list *LinkedList[T]) GetHead() (T, bool) {
	if list.Head != nil {
		return list.Head.Data, true
	}
	return generics.Zero[T](), false
}

func (list *LinkedList[T]) GetTail() (T, bool) {
	if list.Tail != nil {
		return list.Tail.Data, true
	}
	return generics.Zero[T](), false
}

func (list *LinkedList[T]) GetList() []T {
	a := make([]T, 0)
	head := list.Head
	for head != nil {
		a = append(a, head.Data)
		head = head.Next
	}
	return a
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}
