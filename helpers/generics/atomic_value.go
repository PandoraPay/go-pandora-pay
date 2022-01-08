package generics

import "sync/atomic"

type Value[T any] struct {
	v atomic.Value
}

func (v *Value[T]) Load() (z T) {
	out := v.v.Load()
	if out != nil {
		z = out.(T)
	}
	return
}

func (v *Value[T]) Store(val T) {
	v.v.Store(val)
}

func (v *Value[T]) Swap(new T) (old T) {
	return v.v.Swap(new).(T)
}

func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return v.v.CompareAndSwap(old, new)
}
