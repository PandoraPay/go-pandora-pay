package helpers

import "errors"

func ConvertRecoverError(r interface{}) (err error) {
	switch x := r.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New("Unknown panic")
	}
	return
}
