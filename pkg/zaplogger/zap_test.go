package zaplogger

import (
	"os"
	"testing"

	"github.com/fastschema/fastschema/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewZapLogger(t *testing.T) {
	// Create a temporary log file
	logFile := t.TempDir() + "/zaplogger.log"
	defer os.Remove(logFile)

	config := &logger.Config{
		LogFile:    logFile,
		CallerSkip: 1,
	}

	logger, err := NewZapLogger(config)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Test log file creation
	_, err = os.Stat(logFile)
	assert.NoError(t, err)

	// Test logger configuration
	assert.NotNil(t, logger.Logger)
	assert.NotNil(t, logger.LogContext)
	assert.Equal(t, config, logger.config)
}

func TestZapLoggerWithContext(t *testing.T) {
	zapLogger, err := NewZapLogger(&logger.Config{})
	assert.NoError(t, err)
	loggerWithContext := zapLogger.WithContext(logger.LogContext{"key": "value"})
	zapLoggerWithContext, ok := loggerWithContext.(*ZapLogger)
	assert.True(t, ok)
	assert.NotEqual(t, zapLogger, loggerWithContext)
	assert.Equal(t, logger.LogContext{"key": "value"}, zapLoggerWithContext.LogContext)
}

func createLogger(t *testing.T) *ZapLogger {
	logFile := t.TempDir() + "/zaplogger.log"
	logger, _ := NewZapLogger(&logger.Config{LogFile: logFile, DisableConsole: true})
	return logger
}

func TestZapLoggerDebug(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Debug log
	logger.Debug("Debug log")
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"debug"`)
	assert.Contains(t, contentString, `"msg":"Debug log"`)
}

func TestZapLoggerInfo(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Info log
	logger.Info("Info log", "p1", "p2", "p3", "p4")
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"info"`)
	assert.Contains(t, contentString, `"msg":"Info log"`)
	assert.Contains(t, contentString, `"params":["p1","p2","p3","p4"]`)
}

func TestZapLoggerInfof(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Infof log
	logger.Infof("Infof log with params: %s, %d", "param1", 2)
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"info"`)
	assert.Contains(t, contentString, `"msg":"Infof log with params: param1, 2"`)
}

func TestZapLoggerWarn(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Warn log
	logger.Warn("Warn log", "p1", "p2", "p3", "p4")
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"warn"`)
	assert.Contains(t, contentString, `"msg":"Warn log"`)
	assert.Contains(t, contentString, `"params":["p1","p2","p3","p4"]`)
}

func TestZapLoggerError(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Error log
	logger.Error("Error log", "p1", "p2", "p3", "p4")
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"error"`)
	assert.Contains(t, contentString, `"msg":"Error log"`)
	assert.Contains(t, contentString, `"params":["p1","p2","p3","p4"]`)
}

func TestZapLoggerErrorf(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Errorf log
	logger.Errorf("Errorf log with params: %s, %d", "param1", 2)
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"error"`)
	assert.Contains(t, contentString, `"msg":"Errorf log with params: param1, 2"`)
}

func TestZapLoggerDPanic(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test DPanic log
	logger.DPanic("DPanic log", "p1", "p2", "p3", "p4")
	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"dpanic"`)
	assert.Contains(t, contentString, `"msg":"DPanic log"`)
	assert.Contains(t, contentString, `"params":["p1","p2","p3","p4"]`)
}

func TestZapLoggerPanic(t *testing.T) {
	logger := createLogger(t)
	defer os.Remove(logger.config.LogFile)

	// Test Panic log
	assert.Panics(t, func() {
		logger.Panic("Panic log", "p1", "p2", "p3", "p4")
	})

	fileContent, err := os.ReadFile(logger.config.LogFile)
	assert.NoError(t, err)
	contentString := string(fileContent)
	assert.Contains(t, contentString, `"level":"panic"`)
	assert.Contains(t, contentString, `"msg":"Panic log"`)
	assert.Contains(t, contentString, `"params":["p1","p2","p3","p4"]`)
}

func TestGetZapFields(t *testing.T) {
	// Test with empty context
	fields := getZapFields()
	assert.Empty(t, fields)

	// Test with single context
	context := logger.LogContext{"key1": "value1", "key2": "value2"}
	expectedFields := []zapcore.Field{
		zap.Any("key1", "value1"),
		zap.Any("key2", "value2"),
	}
	fields = getZapFields(context)
	assert.ElementsMatch(t, expectedFields, fields)

	// Test with multiple contexts
	contexts := []logger.LogContext{
		{"key1": "value1"},
		{"key2": "value2"},
		{"key3": "value3"},
	}
	expectedFields = []zapcore.Field{
		zap.Any("key1", "value1"),
		zap.Any("key2", "value2"),
		zap.Any("key3", "value3"),
	}
	fields = getZapFields(contexts...)
	assert.ElementsMatch(t, expectedFields, fields)
}
