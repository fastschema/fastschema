package zaplogger

import (
	"errors"
	"testing"

	"github.com/fastschema/fastschema/logger"
	"github.com/stretchr/testify/assert"
)

func TestZapLoggerGetLogContext(t *testing.T) {
	zapLogger := &ZapLogger{}

	t.Run("EmptyParams", func(t *testing.T) {
		msg, contexts := zapLogger.getLogContext()
		assert.Equal(t, "", msg)
		assert.Equal(t, []logger.LogContext{}, contexts)
	})

	t.Run("StringParam", func(t *testing.T) {
		var logContexts []logger.LogContext
		params := []any{"message"}
		msg, contexts := zapLogger.getLogContext(params...)
		assert.Equal(t, "message", msg)
		assert.Equal(t, logContexts, contexts)
	})

	t.Run("ErrorParam", func(t *testing.T) {
		var logContexts []logger.LogContext
		params := []any{errors.New("error message")}
		msg, contexts := zapLogger.getLogContext(params...)
		assert.Equal(t, "error message", msg)
		assert.Equal(t, logContexts, contexts)
	})

	t.Run("ParamsWithLogContext", func(t *testing.T) {
		params := []any{"message", "param1", "param2"}
		msg, contexts := zapLogger.getLogContext(params...)
		assert.Equal(t, "message", msg)
		expectedContexts := []logger.LogContext{
			{"params": []any{"param1", "param2"}},
		}
		assert.Equal(t, expectedContexts, contexts)
	})
}
