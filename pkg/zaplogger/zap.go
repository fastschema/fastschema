package zaplogger

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fastschema/fastschema/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapConfig struct {
	Development bool `json:"development"`
	LogFile     string
	CallerSkip  int
}

type ZapLogger struct {
	*zap.Logger
	logger.Context
	config *ZapConfig
}

func NewZapLogger(config *ZapConfig) (_ *ZapLogger, err error) {
	if config.LogFile != "" {
		if err := os.MkdirAll(path.Dir(config.LogFile), 0755); err != nil {
			return nil, err
		}

		logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		logFile.Close()
	}

	zapConfig := zap.NewProductionEncoderConfig()
	// zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	fileEncoder := zapcore.NewJSONEncoder(zapConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(zapConfig)
	logFile, _ := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	callerSkip := 1
	if config.CallerSkip > 0 {
		callerSkip = config.CallerSkip
	}

	zapLogger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(callerSkip),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	defer func() {
		if e := zapLogger.Sync(); e != nil {
			fmt.Printf("zapLogger.Sync() error: %v\n", e)
		}
	}()

	return &ZapLogger{
		Logger:  zapLogger,
		Context: logger.Context{},
		config:  config,
	}, nil
}

func (l *ZapLogger) WithContext(context logger.Context) logger.Logger {
	return &ZapLogger{
		Logger:  l.Logger.WithOptions(zap.AddCallerSkip(1)),
		Context: context,
		config:  l.config,
	}
}

func (l *ZapLogger) Debug(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.Debug(msg, getZapFields(contexts...)...)
}

func (l *ZapLogger) Info(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.Info(msg, getZapFields(contexts...)...)
}

func (l *ZapLogger) Infof(msg string, params ...any) {
	msg = fmt.Sprintf(msg, params...)
	l.Info(msg)
}

func (l *ZapLogger) Warn(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.Warn(msg, getZapFields(contexts...)...)
}

func (l *ZapLogger) Error(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.Error(msg, getZapFields(contexts...)...)
}

func (l *ZapLogger) Errorf(msg string, params ...any) {
	msg = fmt.Sprintf(msg, params...)
	l.Error(msg)
}

func (l *ZapLogger) DPanic(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.DPanic(msg, getZapFields(contexts...)...)
}

func (l *ZapLogger) Panic(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.Panic(msg, getZapFields(contexts...)...)
}

func (l *ZapLogger) Fatal(params ...any) {
	msg, contexts := l.getLogContext(params...)
	l.Logger.Fatal(msg, getZapFields(contexts...)...)
}

func getZapFields(contexts ...logger.Context) []zapcore.Field {
	var contextFields []zapcore.Field
	for _, context := range contexts {
		for key, val := range context {
			keyIndex := key
			contextFields = append(contextFields, zap.Any(keyIndex, val))
		}
	}
	return contextFields
}
