package app_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/stretchr/testify/assert"
)

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
