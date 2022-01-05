package generics

// usage: Zero(T)()
// e.g. Zero(string)() == ""
func Zero[T any]() T {
	var z T
	return z
}
