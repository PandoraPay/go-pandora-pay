package generics

import (
	"constraints"
	"encoding/json"
)

// usage: Zero(T)()
// e.g. Zero(string)() == ""
func Zero[T any]() (z T) {
	return
}

func Clone[T any](a, z T) (T, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return z, err
	}

	err = json.Unmarshal(data, z)
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
