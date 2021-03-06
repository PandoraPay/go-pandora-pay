package helpers

import "errors"

func ConvertRecoverError(r interface{}) (err error) {
	if r == nil {
		return
	}

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
