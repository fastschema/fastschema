package restresolver

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/fastschema/fastschema/app"
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
	logContext := app.LogContext{
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

func MiddlewareCookie(c *Context) error {
	if c.Cookie("UUID") == "" {
		exp := time.Now().Add(time.Hour * 100 * 365 * 24)
		c.Cookie("UUID", &Cookie{
			Name:     "UUID",
			Value:    uuid.NewString(),
			Expires:  exp,
			HTTPOnly: false,
			SameSite: "lax",
			Secure:   true,
		})
	}
	return c.Next()
}

func MiddlewareCors(c *Context) error {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if c.Method() == "OPTIONS" {
		c.Status(200)
		return nil
	}

	return c.Next()
}

func MiddlewareRecover(c *Context) error {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			msg := fmt.Sprintf("%v %s\n", err, stack[:length])
			c.Logger().Error(msg, app.LogContext{"recovered": true})
			if err := c.Status(http.StatusBadRequest).JSON(app.Map{"error": err.Error()}); err != nil {
				c.Logger().Error(err)
			}
		}
	}()

	return c.Next()
}
