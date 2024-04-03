package restresolver

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/logger"
	"github.com/gofiber/fiber/v2"
)

func transformHandlers(
	r *app.Resource,
	handlers []Handler,
	l logger.Logger,
) []fiber.Handler {
	var fiberHandlers []fiber.Handler

	for i := 0; i < len(handlers); i++ {
		func(r *app.Resource, i int) {
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return handlers[i](createContext(r, c, l))
			})
		}(r, i)
	}

	return fiberHandlers
}

func createContext(r *app.Resource, c *fiber.Ctx, logger logger.Logger) *Context {
	args := make(map[string]string)
	allParams := c.AllParams()
	queries := c.Queries()

	for k, v := range allParams {
		args[k] = v
	}

	for k, v := range queries {
		if _, exists := args[k]; !exists {
			args[k] = v
		}
	}

	return &Context{
		Ctx:      c,
		args:     args,
		resource: r,
		logger:   logger,
	}
}
