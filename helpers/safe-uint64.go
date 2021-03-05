package helpers

import (
	"math"
)

func SafeUint64Add(a *uint64, value uint64) {
	if math.MaxUint64-*a <= value {
		panic("Uint64 would exceed")
	}
	*a += value
}

func SafeUint64Sub(a *uint64, value uint64) {
	if *a < value {
		panic("Uint64 would become negative")
	}
	*a -= value
}

func SafeUint64Update(sign bool, a *uint64, value uint64) {
	if sign {
		if math.MaxUint64-*a <= value {
			panic("Uint64 would exceed")
		}
		*a += value
	} else {
		if *a < value {
			panic("Uint64 would become negative")
		}
		*a -= value
	}
}

func SafeMapUint64Add(m map[string]uint64, key string, value uint64) {
	a := m[key]
	SafeUint64Add(&a, value)
	m[key] = a
}

func SafeMapUint64Sub(m map[string]uint64, key string, value uint64) {
	a := m[key]
	SafeUint64Sub(&a, value)
	m[key] = a
}
