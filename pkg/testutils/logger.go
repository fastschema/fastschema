package testutils

import (
	"fmt"

	"github.com/fastschema/fastschema/logger"
)

type MockLoggerMessage struct {
	Type   string
	Params []any
}

type MockLogger struct {
	Silence  bool
	Messages []*MockLoggerMessage
}

func (l *MockLogger) WithContext(context logger.Context) logger.Logger {
	return l
}

func (l *MockLogger) Last() MockLoggerMessage {
	if len(l.Messages) == 0 {
		return MockLoggerMessage{}
	}
	return *l.Messages[len(l.Messages)-1]
}

func (l *MockLogger) Info(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "Info", Params: params})
}

func (l *MockLogger) Infof(msg string, params ...any) {
	msg = fmt.Sprintf(msg, params...)
	l.Info(msg)
}

func (l *MockLogger) Debug(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "Debug", Params: params})
}
func (l *MockLogger) Warn(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "Warn", Params: params})
}
func (l *MockLogger) Error(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "Error", Params: params})
}
func (l *MockLogger) Errorf(msg string, params ...any) {
	msg = fmt.Sprintf(msg, params...)
	l.Error(msg)
}

func (l *MockLogger) DPanic(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "DPanic", Params: params})
}
func (l *MockLogger) Panic(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "Panic", Params: params})
}
func (l *MockLogger) Fatal(params ...any) {
	if !l.Silence {
		fmt.Println(params...)
	}
	l.Messages = append(l.Messages, &MockLoggerMessage{Type: "Fatal", Params: params})
}
