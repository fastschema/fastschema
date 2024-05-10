package restresolver

import (
	"github.com/fastschema/fastschema/app"
	"github.com/gofiber/fiber/v2"
)

func TransformHandlers(
	r *app.Resource,
	handlers []Handler,
	l app.Logger,
) []fiber.Handler {
	var fiberHandlers []fiber.Handler

	for i := 0; i < len(handlers); i++ {
		func(r *app.Resource, i int) {
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return handlers[i](CreateContext(r, c, l))
			})
		}(r, i)
	}

	return fiberHandlers
}

func CreateContext(r *app.Resource, c *fiber.Ctx, logger app.Logger) *Context {
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

func GetHandlerInfo(handler Handler, logger app.Logger, resources ...*app.Resource) (string, []fiber.Handler) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}

	handlers := TransformHandlers(r, []Handler{handler}, logger)

	return name, handlers
}
