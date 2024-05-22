package restresolver

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/gofiber/fiber/v2"
)

func TransformHandlers(
	r *fs.Resource,
	handlers []Handler,
	l logger.Logger,
) []fiber.Handler {
	var fiberHandlers []fiber.Handler

	for i := 0; i < len(handlers); i++ {
		func(r *fs.Resource, i int) {
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return handlers[i](CreateContext(r, c, l))
			})
		}(r, i)
	}

	return fiberHandlers
}

func CreateContext(r *fs.Resource, c *fiber.Ctx, logger logger.Logger) *Context {
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

func GetHandlerInfo(handler Handler, logger logger.Logger, resources ...*fs.Resource) (string, []fiber.Handler) {
	var r *fs.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}

	handlers := TransformHandlers(r, []Handler{handler}, logger)

	return name, handlers
}
