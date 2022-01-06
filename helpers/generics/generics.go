package generics

import "encoding/json"

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
