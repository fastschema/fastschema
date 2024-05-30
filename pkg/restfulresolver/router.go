package restfulresolver

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"

	"github.com/gofiber/fiber/v2"
)

type Router struct {
	*fiber.App
	fiberGroup *fiber.Group
	logger     logger.Logger
}

func (g *Router) Use(handlers ...Handler) {
	middlewares := []any{}
	transformedHandlers := TransformHandlers(nil, handlers, g.logger)
	for _, handler := range transformedHandlers {
		middlewares = append(middlewares, handler)
	}

	g.fiberGroup.Use(middlewares...)
}

func (g *Router) Group(prefix string, r *fs.Resource, handlers ...Handler) *Router {
	var fiberHandlers []fiber.Handler

	for _, handler := range handlers {
		fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
			return handler(CreateContext(r, c, g.logger))
		})
	}

	gg := g.fiberGroup.Group(prefix, fiberHandlers...).(*fiber.Group)

	return &Router{
		fiberGroup: gg,
		App:        g.App,
		logger:     g.logger,
	}
}

func (g *Router) Get(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Get(path, handlers...).Name(name)
}

func (g *Router) Head(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Head(path, handlers...).Name(name)
}

func (g *Router) Post(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Post(path, handlers...).Name(name)
}

func (g *Router) Put(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Put(path, handlers...).Name(name)
}

func (g *Router) Delete(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Delete(path, handlers...).Name(name)
}

func (g *Router) Connect(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Connect(path, handlers...).Name(name)
}

func (g *Router) Options(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Options(path, handlers...).Name(name)
}

func (g *Router) Trace(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Trace(path, handlers...).Name(name)
}

func (g *Router) Patch(path string, handler Handler, resources ...*fs.Resource) {
	name, handlers := GetHandlerInfo(handler, g.logger, resources...)
	g.fiberGroup.Patch(path, handlers...).Name(name)
}
