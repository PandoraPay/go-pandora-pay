package linked_list

//inspired from https://leetcode.com/problems/sort-list/solution/
func (list *LinkedList[T]) SortList(cmp func(a, b T) bool) {
	list.Head = sortList(list.Head, cmp)

	//TODO: can get a better solution ?
	list.Tail = list.Head
	for list.Tail.Next != nil {
		list.Tail = list.Tail.Next
	}
}

func sortList[T any](head *linkedListItem[T], cmp func(a, b T) bool) *linkedListItem[T] {
	if head == nil || head.Next == nil {
		return head
	}
	mid := getMid(head)
	left := sortList(head, cmp)
	right := sortList(mid, cmp)
	return merge(left, right, cmp)
}

func merge[T any](list1, list2 *linkedListItem[T], cmp func(a, b T) bool) *linkedListItem[T] {
	dummyHead := &linkedListItem[T]{}
	tail := dummyHead
	for list1 != nil && list2 != nil {
		if cmp(list1.Data, list2.Data) {
			tail.Next = list1
			list1 = list1.Next
			tail = tail.Next
		} else {
			tail.Next = list2
			list2 = list2.Next
			tail = tail.Next
		}
	}
	if list1 != nil {
		tail.Next = list1
	} else {
		tail.Next = list2
	}
	return dummyHead.Next
}

func getMid[T any](head *linkedListItem[T]) *linkedListItem[T] {
	var midPrev *linkedListItem[T]
	for head != nil && head.Next != nil {
		if midPrev == nil {
			midPrev = head
		} else {
			midPrev = midPrev.Next
		}
		head = head.Next.Next
	}
	mid := midPrev.Next
	midPrev.Next = nil
	return mid
}
