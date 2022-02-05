package linked_list

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestLinkedList_SortList_smaller(t *testing.T) {
	count := 1000

	list := NewLinkedList[int]()
	for i := 0; i < count; i++ {
		list.Push(rand.Int())
	}
	list.SortList(func(a, b int) bool {
		return a < b
	})
	data := list.GetList()
	assert.Equal(t, count, len(data), "should count")
	for i := 0; i < len(data)-1; i++ {
		assert.Equal(t, true, data[i] < data[i+1], "should be smaller")
	}

}

func TestLinkedList_SortList_greater(t *testing.T) {
	count := 1000

	list := NewLinkedList[int]()
	for i := 0; i < count; i++ {
		list.Push(rand.Int())
	}
	list.SortList(func(a, b int) bool {
		return a > b
	})
	data := list.GetList()
	assert.Equal(t, count, len(data), "should count")
	for i := 0; i < len(data)-1; i++ {
		assert.Equal(t, true, data[i] > data[i+1], "should be greater")
	}

}
