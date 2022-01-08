package container_list

import (
	"pandora-pay/helpers/generics"
	"sync"
)

type ContainerList[T comparable] struct {
	list *generics.Value[[]T]
	lock *sync.RWMutex
}

func (c *ContainerList[T]) Get() []T {
	return c.list.Load()
}

func (c *ContainerList[T]) Push(el T) T {
	c.lock.Lock()
	c.list.Store(append(c.list.Load(), el))
	c.lock.Unlock()
	return el
}

func (c *ContainerList[T]) Remove(el T) bool {

	c.lock.Lock()
	defer c.lock.Unlock()

	all := c.list.Load()
	for i, el2 := range all {
		if el2 == el {
			//removing from array array
			list2 := make([]T, len(all)-1)
			copy(list2, all)
			if len(all) > 1 && i != len(all)-1 {
				list2[i] = all[len(all)-1]
			}
			c.list.Store(list2)
			return true
		}
	}
	return false
}

func (c *ContainerList[T]) RemoveAll() []T {
	c.lock.Lock()
	defer c.lock.Unlock()

	list := c.list.Load()
	c.list.Store(make([]T, 0))
	return list
}

func NewContainerList[T comparable]() *ContainerList[T] {
	c := &ContainerList[T]{
		&generics.Value[[]T]{},
		&sync.RWMutex{},
	}
	c.list.Store(make([]T, 0))
	return c
}
