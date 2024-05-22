package db_test

import (
	"errors"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFound(t *testing.T) {
	// Test when err is nil
	assert.False(t, db.IsNotFound(nil))

	// Test when err is not a NotFoundError
	err := errors.New("some error")
	assert.False(t, db.IsNotFound(err))

	// Test when err is a NotFoundError
	notFoundErr := &db.NotFoundError{}
	assert.True(t, db.IsNotFound(notFoundErr))
}

func TestNotFoundErrorError(t *testing.T) {
	// Test Error() method of NotFoundError
	err := &db.NotFoundError{Message: "not found"}
	assert.Equal(t, "not found", err.Error())
}
