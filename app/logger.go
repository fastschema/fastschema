package app

import "fmt"

type LogContext = map[string]any

type Logger interface {
	Info(...any)
	Infof(string, ...any)
	Error(...any)
	Errorf(string, ...any)
	Debug(...any)
	Fatal(...any)
	Warn(...any)
	Panic(...any)
	DPanic(...any)
	WithContext(context LogContext, callerSkips ...int) Logger
}

type MockLoggerMessage struct {
	Type   string `json:"type"`
	Params []any  `json:"params"`
}

func (m MockLoggerMessage) String() string {
	return fmt.Sprintf("%s: %v", m.Type, m.Params)
}

type MockLogger struct {
	Silence  bool
	Messages []*MockLoggerMessage
}

func CreateMockLogger(silences ...bool) *MockLogger {
	silences = append(silences, false)
	mockLogger := &MockLogger{
		Silence: silences[0],
	}

	return mockLogger
}

func (l *MockLogger) WithContext(context LogContext, callerSkips ...int) Logger {
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
