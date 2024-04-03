package restresolver

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/fastschema/fastschema/logger"
)

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
			c.Logger().Error(msg, logger.Context{"recovered": true})
			if err := c.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()}); err != nil {
				c.Logger().Error(err)
			}
		}
	}()

	return c.Next()
}
