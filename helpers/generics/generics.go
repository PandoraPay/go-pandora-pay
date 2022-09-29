package generics

import (
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/constraints"
)

// usage: Zero(T)()
// e.g. Zero(string)() == ""
func Zero[T any]() (z T) {
	return
}

func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

func Clone[T any](a, z T) (T, error) {
	data, err := msgpack.Marshal(a)
	if err != nil {
		return z, err
	}

	err = msgpack.Unmarshal(data, z)
	return z, err
}

func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}
