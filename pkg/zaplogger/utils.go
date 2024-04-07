package zaplogger

import (
	"strconv"

	"github.com/fastschema/fastschema/app"
)

func (l *ZapLogger) getLogContext(params ...any) (string, []app.LogContext) {
	if len(params) == 0 {
		return "", []app.LogContext{}
	}

	msg := ""
	ctx := app.LogContext{}
	contexts := []app.LogContext{l.LogContext}

	if m, ok := params[0].(string); ok {
		msg = m
		params = params[1:]
	}

	for i, p := range params {
		if err, ok := p.(error); ok && msg == "" {
			msg = err.Error()
		}

		if c, ok := p.(app.LogContext); ok {
			contexts = append(contexts, c)
		} else {
			ctx[strconv.Itoa(i)] = p
		}
	}

	contexts = append(contexts, ctx)

	return msg, contexts
}
