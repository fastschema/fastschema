package logger

type Context map[string]any

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
	WithContext(context Context) Logger
}
