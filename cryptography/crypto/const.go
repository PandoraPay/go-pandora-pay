package crypto

import (
	"errors"
	"fmt"
)

const POINT_SIZE = 33        // this can be optimized to 33 bytes
const FIELDELEMENT_SIZE = 32 // why not have bigger curves

// protocol supports amounts upto this amounts
const MAX_AMOUNT = 18446744073709551616 // 2^64 - 1,

const PROTOCOL_CONSTANT = "PANDORA"

// checks a number is power of 2
func IsPowerOf2(num int) bool {
	for num >= 2 {
		if num%2 != 0 {
			return false
		}
		num = num / 2
	}
	return num == 1
}

// tell what power a number is
func GetPowerof2(num int) (int, error) {

	if num <= 0 {
		return 0, errors.New("number cannot be less than 0")
	}

	if !IsPowerOf2(num) {
		return 0, errors.New(fmt.Sprintf("number(%d) must be power of 2", num))
	}

	power := 0
	calculated := 1
	for ; calculated != num; power++ {
		calculated *= 2
	}
	return power, nil
}
