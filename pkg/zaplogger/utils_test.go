package zaplogger

import (
	"errors"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/stretchr/testify/assert"
)

func TestZapLoggerGetLogContext(t *testing.T) {
	logger := &ZapLogger{}

	t.Run("EmptyParams", func(t *testing.T) {
		msg, contexts := logger.getLogContext()
		assert.Equal(t, "", msg)
		assert.Equal(t, []app.LogContext{}, contexts)
	})

	t.Run("StringParam", func(t *testing.T) {
		var logContexts []app.LogContext
		params := []any{"message"}
		msg, contexts := logger.getLogContext(params...)
		assert.Equal(t, "message", msg)
		assert.Equal(t, logContexts, contexts)
	})

	t.Run("ErrorParam", func(t *testing.T) {
		var logContexts []app.LogContext
		params := []any{errors.New("error message")}
		msg, contexts := logger.getLogContext(params...)
		assert.Equal(t, "error message", msg)
		assert.Equal(t, logContexts, contexts)
	})

	t.Run("ParamsWithLogContext", func(t *testing.T) {
		params := []any{"message", "param1", "param2"}
		msg, contexts := logger.getLogContext(params...)
		assert.Equal(t, "message", msg)
		expectedContexts := []app.LogContext{
			{"params": []any{"param1", "param2"}},
		}
		assert.Equal(t, expectedContexts, contexts)
	})
}
