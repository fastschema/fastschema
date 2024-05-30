package logger_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/logger"
	"github.com/stretchr/testify/assert"
)

func TestCreateMockLogger(t *testing.T) {
	// Test case 1: Create mock l with default silence value
	l := logger.CreateMockLogger()
	assert.NotNil(t, l)
	assert.False(t, l.Silence)

	// Test case 2: Create mock logger with silence value set to true
	l = logger.CreateMockLogger(true)
	assert.NotNil(t, l)
	assert.True(t, l.Silence)

	// Test case 3: Create mock logger with multiple silence values
	l = logger.CreateMockLogger(true, false, true)
	assert.NotNil(t, l)
	assert.True(t, l.Silence)
}

func TestMockLoggerWithContext(t *testing.T) {
	// Test case 1: WithContext should return the same l instance
	l := logger.CreateMockLogger()
	context := logger.LogContext{}
	result := l.WithContext(context)
	assert.Equal(t, l, result)

	// Test case 2: WithContext should not modify the original logger
	l = logger.CreateMockLogger()
	context = logger.LogContext{}
	l.WithContext(context)
	assert.False(t, l.Silence)

	// Test case 3: WithContext should not modify the provided context
	l = logger.CreateMockLogger()
	context = logger.LogContext{"extra": "debug"}
	assert.Equal(t, l, l.WithContext(context))
}

func TestMockLoggerMethods(t *testing.T) {
	l := logger.CreateMockLogger()
	methodsMap := map[string]func(params ...any){
		"Info":   l.Info,
		"Error":  l.Error,
		"Debug":  l.Debug,
		"Warn":   l.Warn,
		"Panic":  l.Panic,
		"DPanic": l.DPanic,
		"Fatal":  l.Fatal,
	}

	for method, fn := range methodsMap {
		fn("test")
		assert.Equal(t, method, l.Last().Type)
		assert.Equal(t, []any{"test"}, l.Last().Params)
		assert.Equal(t, fmt.Sprintf(`%s: [test]`, method), l.Last().String())
	}

	l.Infof("test %s", "message")
	assert.Equal(t, "Info", l.Last().Type)
	assert.Equal(t, []any{"test message"}, l.Last().Params)
	assert.Equal(t, "Info: [test message]", l.Last().String())

	l.Errorf("test %s", "message")
	assert.Equal(t, "Error", l.Last().Type)
	assert.Equal(t, []any{"test message"}, l.Last().Params)
	assert.Equal(t, "Error: [test message]", l.Last().String())
}

func TestMockLoggerMessages(t *testing.T) {
	l := logger.CreateMockLogger()
	assert.Equal(t, logger.MockLoggerMessage{}, l.Last())

	l.Info("test")
	l.Error("test")
	l.Debug("test")

	assert.Equal(t, 3, len(l.Messages))
	assert.Equal(t, "Info", l.Messages[0].Type)
	assert.Equal(t, "Error", l.Messages[1].Type)
	assert.Equal(t, "Debug", l.Messages[2].Type)
}
