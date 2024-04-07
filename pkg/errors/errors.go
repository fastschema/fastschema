package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/xerrors"
)

func CreateErrorFn(status int) func(msgs ...any) *Error {
	return func(msgs ...any) *Error {
		err := &Error{Status: status}
		if len(msgs) > 0 {
			format, ok := msgs[0].(string)
			if ok {
				err.Message = fmt.Sprintf(format, msgs[1:]...)
			} else {
				err.Message = fmt.Sprint(msgs...)
			}
		}

		return err
	}
}

var (
	InternalServerError = CreateErrorFn(http.StatusInternalServerError)
	Unauthorized        = CreateErrorFn(http.StatusUnauthorized)
	Unauthenticated     = CreateErrorFn(http.StatusUnauthorized)
	BadRequest          = CreateErrorFn(http.StatusBadRequest)
	Forbidden           = CreateErrorFn(http.StatusForbidden)
	NotFound            = CreateErrorFn(http.StatusNotFound)
	BadGateway          = CreateErrorFn(http.StatusBadGateway)
	UnprocessableEntity = CreateErrorFn(http.StatusUnprocessableEntity)
)

var errStatusMap = map[int]func(msgs ...any) *Error{
	http.StatusInternalServerError: InternalServerError,
	http.StatusUnauthorized:        Unauthorized,
	http.StatusBadRequest:          BadRequest,
	http.StatusForbidden:           Forbidden,
	http.StatusNotFound:            NotFound,
	http.StatusBadGateway:          BadGateway,
	http.StatusUnprocessableEntity: UnprocessableEntity,
}

func GetErrorByStatus(status int, err error) *Error {
	newErrorFn, ok := errStatusMap[status]
	if ok {
		return newErrorFn(err.Error())
	}

	return InternalServerError(err.Error())
}

// Error is error object.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`

	err   error
	base  error
	frame xerrors.Frame
}

// New creates new error object. Example: errors.New("Login failed", "LOGIN_FAILED", "Invalid username or password")
//
//	message: error message
//	extras: error detail and error code
func New(message string, extras ...string) *Error {
	const skip = 1

	e := &Error{
		Message: message,
		frame:   xerrors.Caller(skip),
	}

	if len(extras) > 0 {
		e.Detail = extras[0]
	}

	if len(extras) > 1 {
		e.Code = extras[1]
	}

	return e
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	var e, t *Error
	if errors.As(err, &e) {
		if errors.As(target, &t) {
			return errors.Is(e.base, t.base)
		}
		return errors.Is(e.base, target)
	}

	if errors.As(target, &t) {
		return errors.Is(err, t.base)
	}
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
func As(err error, target any) bool {
	return errors.As(err, target)
}

// From is constructor to create error object.
func From(err error) *Error {
	var e *Error

	switch {
	case errors.As(err, &e):
		return e
	default:
		errors.As(InternalServerError().Wrap(err), &e)
		return e
	}
}

// Error is error interface implementation.
func (e *Error) Error() string {
	result := ""
	message := e.Message
	code := e.Code

	if e.Status != 0 {
		if message == "" {
			message = http.StatusText(e.Status)
		}

		if code == "" {
			code = strconv.Itoa(e.Status)
		}
	}

	if code != "" {
		result += fmt.Sprintf("[%s] ", code)
	}

	result += message

	if e.Detail != "" {
		result += ": " + e.Detail
	}

	return result
}

// Format is error interface implementation.
func (e *Error) Format(f fmt.State, c rune) {
	xerrors.FormatError(e, f, c)
}

// FormatError is error interface implementation.
func (e *Error) FormatError(p xerrors.Printer) error {
	e.frame.Format(p)

	return e.Unwrap()
}

func (e *Error) Messagef(format string, a ...any) *Error {
	e.Message = fmt.Sprintf(format, a...)

	return e
}

// Unwrap returns the result of calling the Unwrap method on err.
func (e *Error) Unwrap() error {
	return e.err
}

// Wrap is constructor to wrap error object.
func (e *Error) Wrap(target error) error {
	const skip = 1
	e.Message = target.Error()
	err := *e
	err.base = e
	err.err = target
	err.frame = xerrors.Caller(skip)

	return &err
}

// Trace is constructor to wrap error object.
func (e *Error) Trace() error {
	const skip = 1
	err := *e
	err.base = e
	err.frame = xerrors.Caller(skip)

	return &err
}

// MarshalJSON is json.Marshaler interface implementation.
func (e *Error) MarshalJSON() ([]byte, error) {
	message := e.Message
	code := e.Code

	if e.Status != 0 {
		if message == "" {
			message = http.StatusText(e.Status)
		}

		if code == "" {
			code = strconv.Itoa(e.Status)
		}
	}

	return json.Marshal(struct {
		Code    string `json:"code,omitempty"`
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
	}{
		Code:    code,
		Message: message,
		Detail:  e.Detail,
	})
}
