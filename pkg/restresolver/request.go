package restresolver

import (
	"time"

	"github.com/fastschema/fastschema/logger"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-Id"

func MiddlewareRequestID(c *Context) error {
	requestID := c.Header(HeaderRequestID)

	if requestID == "" {
		requestID = uuid.NewString()
	}

	c.Value("request_id", requestID)
	c.Header(HeaderRequestID, requestID)

	return c.Next()
}

func MiddlewareRequestLog(c *Context) error {
	start := time.Now()
	err := c.Next()
	latency := time.Since(start).Round(time.Millisecond)
	logContext := logger.Context{
		"latency": latency.String(),
		"status":  c.Response().StatusCode(),
		"method":  c.Method(),
		"path":    c.Path(),
		"ip":      c.IP(),
	}

	if err != nil {
		logContext["error"] = err.Error()
	}

	c.Logger().Info("Request completed", logContext)
	return err
}
