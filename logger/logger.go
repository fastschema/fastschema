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

// var loggerInstance Logger

// func New(logger Logger) Logger {
// 	loggerInstance = logger
// 	return loggerInstance
// }

// func Get() Logger {
// 	return loggerInstance
// }

// func Debug(params ...any) {
// 	loggerInstance.Debug(params...)
// }

// func Info(params ...any) {
// 	loggerInstance.Info(params...)
// }

// func Infof(format string, params ...any) {
// 	msg := fmt.Sprintf(format, params...)
// 	loggerInstance.Info(msg)
// }

// func Warn(params ...any) {
// 	loggerInstance.Warn(params...)
// }

// func Error(params ...any) {
// 	loggerInstance.Error(params...)
// }

// func DPanic(params ...any) {
// 	loggerInstance.DPanic(params...)
// }

// func Panic(params ...any) {
// 	loggerInstance.Panic(params...)
// }

// func Fatal(params ...any) {
// 	loggerInstance.Fatal(params...)
// }
