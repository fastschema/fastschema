package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateErrorFn(t *testing.T) {
	createError := CreateErrorFn(500)

	t.Run("WithFormatString", func(t *testing.T) {
		err := createError("Error: %s", "Something went wrong")
		assert.NotNil(t, err)
		assert.Equal(t, 500, err.Status)
		assert.Equal(t, "Error: Something went wrong", err.Message)
	})

	t.Run("WithoutFormatString", func(t *testing.T) {
		err := createError("An error occurred")
		assert.NotNil(t, err)
		assert.Equal(t, 500, err.Status)
		assert.Equal(t, "An error occurred", err.Message)
	})

	t.Run("WithoutFormatString2", func(t *testing.T) {
		err := createError(500)
		assert.NotNil(t, err)
		assert.Equal(t, 500, err.Status)
		assert.Equal(t, "500", err.Message)
	})

	t.Run("WithoutMessage", func(t *testing.T) {
		err := createError()
		assert.NotNil(t, err)
		assert.Equal(t, 500, err.Status)
		assert.Equal(t, "", err.Message)
	})
}

func TestGetErrorByStatus(t *testing.T) {
	err := errors.New("test error")

	t.Run("ExistingStatus", func(t *testing.T) {
		expected := "test error"
		result := GetErrorByStatus(400, err)
		assert.NotNil(t, result)
		assert.Equal(t, expected, result.Message)
		assert.Equal(t, 400, result.Status)
	})

	t.Run("NonExistingStatus", func(t *testing.T) {
		expected := "test error"
		result := GetErrorByStatus(1000, err)
		assert.NotNil(t, result)
		assert.Equal(t, expected, result.Message)
		assert.Equal(t, 500, result.Status)
	})
}

func TestNew(t *testing.T) {
	err := New("Something went wrong", "detail", "code")

	assert.NotNil(t, err)
	assert.Equal(t, "Something went wrong", err.Message)
	assert.Equal(t, "detail", err.Detail)
	assert.Equal(t, "code", err.Code)
}

func TestNewWithoutExtras(t *testing.T) {
	err := New("Something went wrong")

	assert.NotNil(t, err)
	assert.Equal(t, "Something went wrong", err.Message)
	assert.Equal(t, "", err.Detail)
	assert.Equal(t, "", err.Code)
}

func TestNewWithDetailOnly(t *testing.T) {
	err := New("Something went wrong", "detail")

	assert.NotNil(t, err)
	assert.Equal(t, "Something went wrong", err.Message)
	assert.Equal(t, "detail", err.Detail)
	assert.Equal(t, "", err.Code)
}

func TestNewWithCodeOnly(t *testing.T) {
	err := New("Something went wrong", "", "code")

	assert.NotNil(t, err)
	assert.Equal(t, "Something went wrong", err.Message)
	assert.Equal(t, "", err.Detail)
	assert.Equal(t, "code", err.Code)
}

func TestIs(t *testing.T) {
	err := New("Something went wrong")
	target := New("Something went wrong")

	t.Run("SameError", func(t *testing.T) {
		result := Is(err, target)
		assert.True(t, result)
	})

	t.Run("DifferentError", func(t *testing.T) {
		otherErr := fmt.Errorf("Another error")
		result := Is(err, otherErr)
		assert.False(t, result)
	})

	e := fmt.Errorf("Another error")
	t.Run("NilError", func(t *testing.T) {
		result := Is(e, target)
		assert.False(t, result)
	})

	t.Run("NilTarget", func(t *testing.T) {
		result := Is(e, nil)
		assert.False(t, result)
	})
}

type MyCustomError struct {
	message string
}

func (e *MyCustomError) Error() string {
	return e.message
}

func TestAs(t *testing.T) {
	err := errors.New("test error")

	t.Run("AsSuccess", func(t *testing.T) {
		target := errors.New("test error")
		result := As(err, &target)
		assert.True(t, result)
		assert.NotNil(t, target)
		assert.Equal(t, "test error", target.Error())
	})

	t.Run("AsFailure", func(t *testing.T) {
		target := &MyCustomError{}
		otherErr := fmt.Errorf("Another error")
		result := As(otherErr, &target)
		assert.False(t, result)
	})
}

func TestFrom(t *testing.T) {
	t.Run("ErrorAsError", func(t *testing.T) {
		err := errors.New("test error")
		result := From(err)
		assert.NotNil(t, result)
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("ErrorAsInternalServerError", func(t *testing.T) {
		e := &Error{}
		internalServerError := InternalServerError()
		result := From(internalServerError)
		assert.True(t, errors.As(result, &e))
		assert.Equal(t, internalServerError, e)
	})
}

func TestError(t *testing.T) {
	t.Run("EmptyError", func(t *testing.T) {
		err := &Error{}
		expected := ""
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageOnly", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
		}
		expected := "Something went wrong"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageAndDetail", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Detail:  "Additional details",
		}
		expected := "Something went wrong: Additional details"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageAndCode", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Code:    "ERROR_CODE",
		}
		expected := "[ERROR_CODE] Something went wrong"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageAndStatus", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Status:  500,
		}
		expected := "[500] Something went wrong"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageAndStatusAndDetail", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Status:  500,
			Detail:  "Additional details",
		}
		expected := "[500] Something went wrong: Additional details"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageAndStatusAndCode", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Status:  500,
			Code:    "ERROR_CODE",
		}
		expected := "[ERROR_CODE] Something went wrong"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorMessageAndStatusAndCodeAndDetail", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Status:  500,
			Code:    "ERROR_CODE",
			Detail:  "Additional details",
		}
		expected := "[ERROR_CODE] Something went wrong: Additional details"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("StatusTextAndCode", func(t *testing.T) {
		err := &Error{
			Status: 404,
			Code:   "ERROR_CODE",
		}
		expected := "[ERROR_CODE] Not Found"
		result := err.Error()
		assert.Equal(t, expected, result)
	})

	t.Run("StatusTextAndCodeAndDetail", func(t *testing.T) {
		err := &Error{
			Status: 404,
			Code:   "ERROR_CODE",
			Detail: "Additional details",
		}
		expected := "[ERROR_CODE] Not Found: Additional details"
		result := err.Error()
		assert.Equal(t, expected, result)
	})
}

func TestErrorMessagef(t *testing.T) {
	err := &Error{
		Message: "Something went wrong",
	}

	t.Run("WithFormatString", func(t *testing.T) {
		result := err.Messagef("Error: %s", "Something went wrong")
		assert.Equal(t, "Error: Something went wrong", result.Message)
	})

	t.Run("WithoutFormatString", func(t *testing.T) {
		result := err.Messagef("An error occurred")
		assert.Equal(t, "An error occurred", result.Message)
	})

	t.Run("WithoutFormatString2", func(t *testing.T) {
		result := err.Messagef("500")
		assert.Equal(t, "500", result.Message)
	})
}

func TestUnwrap(t *testing.T) {
	err := errors.New("test error")
	e := &Error{
		err: err,
	}

	t.Run("Unwrap", func(t *testing.T) {
		result := e.Unwrap()
		assert.Equal(t, err, result)
	})
}

func TestErrorTrace(t *testing.T) {
	err := &Error{
		Message: "Something went wrong",
	}

	t.Run("Trace", func(t *testing.T) {
		result := err.Trace()
		assert.NotNil(t, result)
		assert.Equal(t, err, result.(*Error).base)
	})
}

func TestErrorMarshalJSON(t *testing.T) {
	err := &Error{
		Message: "Something went wrong",
		Code:    "ERROR_CODE",
		Detail:  "Additional details",
	}

	t.Run("WithStatus", func(t *testing.T) {
		err.Status = 500
		expected := `{"code":"ERROR_CODE","message":"Something went wrong","detail":"Additional details"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutStatus", func(t *testing.T) {
		err.Status = 0
		expected := `{"code":"ERROR_CODE","message":"Something went wrong","detail":"Additional details"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutMessage", func(t *testing.T) {
		err := &Error{
			Message: "Something went wrong",
			Code:    "500",
			Detail:  "Additional details",
		}
		err.Message = ""
		expected := `{"code":"500","message":"","detail":"Additional details"}`
		result, e := err.MarshalJSON()
		assert.NoError(t, e)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutCode", func(t *testing.T) {
		err.Code = ""
		err.Status = 500
		expected := `{"code":"500","message":"Something went wrong","detail":"Additional details"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutDetail", func(t *testing.T) {
		err.Detail = ""
		expected := `{"code":"500","message":"Something went wrong"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutMessageAndCode", func(t *testing.T) {
		err.Message = ""
		err.Code = ""
		err.Status = 500
		expected := `{"code":"500","message":"Internal Server Error"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutMessageAndDetail", func(t *testing.T) {
		err.Message = ""
		err.Detail = ""
		err.Status = 500
		expected := `{"code":"500","message":"Internal Server Error"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutCodeAndDetail", func(t *testing.T) {
		err.Code = ""
		err.Detail = ""
		err.Status = 500
		expected := `{"code":"500","message":"Internal Server Error"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})

	t.Run("WithoutMessageAndCodeAndDetail", func(t *testing.T) {
		err.Message = ""
		err.Code = ""
		err.Detail = ""
		err.Status = 500
		expected := `{"code":"500","message":"Internal Server Error"}`
		result, err := err.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, expected, string(result))
	})
}
