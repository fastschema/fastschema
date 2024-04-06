package app_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewPagination(t *testing.T) {
	total := uint(100)
	perPage := uint(10)
	currentPage := uint(1)
	data := []int{1, 2, 3, 4, 5}

	pagination := app.NewPagination(total, perPage, currentPage, data)

	assert.NotNil(t, pagination)
	assert.Equal(t, total, pagination.Pagination.Total)
	assert.Equal(t, perPage, pagination.Pagination.PerPage)
	assert.Equal(t, currentPage, pagination.Pagination.CurrentPage)
	assert.Equal(t, uint(math.Ceil(float64(total)/float64(perPage))), pagination.Pagination.LastPage)
	assert.Equal(t, data, pagination.Data)
}

func TestNewResult(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		data := "test data"
		err := errors.New("test error")
		result := app.NewResult(data, err)

		assert.NotNil(t, result)
		assert.Equal(t, data, result.Data)
		assert.NotNil(t, result.Error)
		assert.Equal(t, err.Error(), result.Error.Error())
	})

	t.Run("with custom error", func(t *testing.T) {
		data := "test data"
		err := fmt.Errorf("test error")
		result := app.NewResult(data, err)

		assert.NotNil(t, result)
		assert.Equal(t, data, result.Data)
		assert.NotNil(t, result.Error)
		assert.Equal(t, "[500] test error", result.Error.Error())
		assert.Equal(t, "", result.Error.Code)
	})

	t.Run("without error", func(t *testing.T) {
		data := "test data"
		result := app.NewResult(data, nil)

		assert.NotNil(t, result)
		assert.Equal(t, data, result.Data)
		assert.Nil(t, result.Error)
	})
}
