package restfulresolver

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type MethodData struct {
	Path    string
	Handler func(path string, handler Handler, resources ...*fs.Resource)
}

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

func TransformMiddlewares(inputs []fs.Middleware) []Handler {
	middlewares := make([]Handler, 0)
	for _, middleware := range inputs {
		middlewares = append(middlewares, func(c *Context) error {
			if err := middleware(c); err != nil {
				var fiberError *fiber.Error
				if errors.As(err, &fiberError) {
					err = errors.GetErrorByStatus(fiberError.Code, err)
				}

				result := fs.NewResult(nil, err)

				if result.Error != nil && result.Error.Status != 0 {
					c.Status(result.Error.Status)
				}

				return c.JSON(result)
			}

			return nil
		})
	}

	return middlewares
}

func CreateContext(r *fs.Resource, c *fiber.Ctx, logger logger.Logger, wsClients ...fs.WSClient) *Context {
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

	ctx := &Context{
		Ctx:        c,
		RequestCtx: c.Context(),
		args:       args,
		resource:   r,
		logger:     logger,
	}

	return ctx
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

func CreateHTTPHandlers(router *Router, meta *fs.Meta) []MethodData {
	return utils.Filter([]MethodData{{
		Path:    meta.Get,
		Handler: router.Get,
	}, {
		Path:    meta.Head,
		Handler: router.Head,
	}, {
		Path:    meta.Post,
		Handler: router.Post,
	}, {
		Path:    meta.Put,
		Handler: router.Put,
	}, {
		Path:    meta.Delete,
		Handler: router.Delete,
	}, {
		Path:    meta.Connect,
		Handler: router.Connect,
	}, {
		Path:    meta.Options,
		Handler: router.Options,
	}, {
		Path:    meta.Trace,
		Handler: router.Trace,
	}, {
		Path:    meta.Patch,
		Handler: router.Patch,
	}}, func(m MethodData) bool {
		return m.Path != ""
	})
}
