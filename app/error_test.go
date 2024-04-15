package app

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNotFound(t *testing.T) {
	// Test when err is nil
	assert.False(t, IsNotFound(nil))

	// Test when err is not a NotFoundError
	err := errors.New("some error")
	assert.False(t, IsNotFound(err))

	// Test when err is a NotFoundError
	notFoundErr := &NotFoundError{}
	assert.True(t, IsNotFound(notFoundErr))
}

func TestNotFoundErrorError(t *testing.T) {
	// Test Error() method of NotFoundError
	err := &NotFoundError{Message: "not found"}
	assert.Equal(t, "not found", err.Error())
}
