package generics

import "sync"

type Map[K any, V any] struct {
	v sync.Map
}

func (m *Map[K, V]) Load(key K) (V, bool) {
	a, b := m.v.Load(key)
	if !b {
		return Zero[V](), false
	}
	return a.(V), b
}

func (m *Map[K, V]) Store(key K, value V) {
	m.v.Store(key, value)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	a, b := m.v.LoadOrStore(key, value)
	return a.(V), b
}

func (m *Map[K, V]) LoadAndDelete(key K) (V, bool) {
	a, b := m.v.LoadAndDelete(key)
	if !b {
		return Zero[V](), false
	}
	return a.(V), b
}

func (m *Map[K, V]) Delete(key K) {
	m.v.Delete(key)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.v.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
