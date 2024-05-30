package db

import "errors"

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e *NotFoundError
	return errors.As(err, &e)
}
