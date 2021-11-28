package helpers

import "errors"

func ReturnErrorIfNot(err error, otherwiseErrorMessage string) error {
	if err != nil {
		return err
	}
	return errors.New(otherwiseErrorMessage)
}
