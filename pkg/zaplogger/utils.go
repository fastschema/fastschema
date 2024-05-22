package zaplogger

import (
	"github.com/fastschema/fastschema/logger"
)

func (l *ZapLogger) getLogContext(params ...any) (string, []logger.LogContext) {
	if len(params) == 0 {
		return "", []logger.LogContext{}
	}

	msg := ""
	var contexts []logger.LogContext

	if l.LogContext != nil {
		contexts = append(contexts, l.LogContext)
	}

	firstParam := params[0]

	if m, ok := firstParam.(string); ok {
		msg = m
		params = params[1:]
	}

	if err, ok := firstParam.(error); ok {
		msg = err.Error()
		params = params[1:]
	}

	if len(params) > 0 {
		contexts = append(contexts, logger.LogContext{"params": params})
	}

	return msg, contexts
}
