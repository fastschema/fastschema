package zaplogger

import (
	"strconv"

	"github.com/fastschema/fastschema/logger"
)

func (l *ZapLogger) getLogContext(params ...any) (string, []logger.Context) {
	if len(params) == 0 {
		return "", []logger.Context{}
	}

	msg := ""
	ctx := logger.Context{}
	contexts := []logger.Context{l.Context}

	if m, ok := params[0].(string); ok {
		msg = m
		params = params[1:]
	}

	for i, p := range params {
		if err, ok := p.(error); ok && msg == "" {
			msg = err.Error()
		}

		if c, ok := p.(logger.Context); ok {
			contexts = append(contexts, c)
		} else {
			ctx[strconv.Itoa(i)] = p
		}
	}

	contexts = append(contexts, ctx)

	return msg, contexts
}
