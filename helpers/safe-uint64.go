package helpers

import (
	"errors"
	"math"
)

func SafeUint64Add(a *uint64, value uint64) error {
	if math.MaxUint64-*a <= value {
		return errors.New("Uint64 would exceed")
	}
	*a += value
	return nil
}

func SafeUint64Sub(a *uint64, value uint64) error {
	if *a < value {
		return errors.New("Uint64 would become negative")
	}
	*a -= value
	return nil
}

func SafeUint64Update(sign bool, a *uint64, value uint64) error {
	if sign {
		if math.MaxUint64-*a <= value {
			return errors.New("Uint64 would exceed")
		}
		*a += value
	} else {
		if *a < value {
			return errors.New("Uint64 would become negative")
		}
		*a -= value
	}
	return nil
}

func SafeMapUint64Add(m map[string]uint64, key string, value uint64) error {
	a := m[key]
	if math.MaxUint64-a <= value {
		return errors.New("Uint64 would exceed")
	}
	m[key] = a + value
	return nil
}

func SafeMapUint64Sub(m map[string]uint64, key string, value uint64) error {
	a := m[key]
	if a < value {
		return errors.New("Uint64 would become negative")
	}
	m[key] = a - value
	return nil
}
