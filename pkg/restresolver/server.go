package restresolver

import (
	"net/http"

	"github.com/fastschema/fastschema/app"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/utils"
)

type Server struct {
	*fiber.App
	logger app.Logger
}

type Config struct {
	AppName     string
	JSONEncoder utils.JSONMarshal
	Logger      app.Logger
}

func New(config Config) *Server {
	app := fiber.New(fiber.Config{
		AppName:               config.AppName,
		StrictRouting:         true,
		CaseSensitive:         true,
		EnablePrintRoutes:     false,
		DisableStartupMessage: true,
		JSONEncoder:           config.JSONEncoder,
		// Prefork:       true,
		// Immutable:     true,
	})

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))

	return &Server{
		App:    app,
		logger: config.Logger,
	}
}

func (s *Server) Test(req *http.Request, msTimeout ...int) (resp *http.Response, err error) {
	return s.App.Test(req, msTimeout...)
}

func (s *Server) Listen(address string) error {
	return s.App.Listen(address)
}

func (s *Server) Use(handlers ...Handler) {
	middlewares := []any{}
	transformedHandlers := TransformHandlers(nil, handlers, s.logger)
	for _, handler := range transformedHandlers {
		middlewares = append(middlewares, handler)
	}

	s.App.Use(middlewares...)
}

func (s *Server) Group(prefix string, r *app.Resource, handlers ...Handler) *Router {
	var fiberHandlers []fiber.Handler

	for _, handler := range handlers {
		fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
			return handler(CreateContext(r, c, s.logger))
		})
	}

	g := s.App.Group(prefix, fiberHandlers...).(*fiber.Group)

	return &Router{
		fiberGroup: g,
		App:        s.App,
		logger:     s.logger,
	}
}

func (s *Server) Static(prefix, root string, configs ...StaticConfig) {
	config := fiber.Static{}

	if len(configs) > 0 {
		config = fiber.Static{
			Index:         configs[0].Index,
			Browse:        configs[0].Browse,
			MaxAge:        configs[0].MaxAge,
			Compress:      configs[0].Compress,
			ByteRange:     configs[0].ByteRange,
			Download:      configs[0].Download,
			CacheDuration: configs[0].CacheDuration,
		}
	}
	s.App.Static(prefix, root, config)
}

func (s *Server) Get(path string, handler Handler, resources ...*app.Resource) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}
	handlers := TransformHandlers(r, []Handler{handler}, s.logger)
	s.App.Get(path, handlers...).Name(name)
}

func (s *Server) Post(path string, handler Handler, resources ...*app.Resource) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}
	handlers := TransformHandlers(r, []Handler{handler}, s.logger)
	s.App.Post(path, handlers...).Name(name)
}

func (s *Server) Put(path string, handler Handler, resources ...*app.Resource) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}
	handlers := TransformHandlers(r, []Handler{handler}, s.logger)
	s.App.Put(path, handlers...).Name(name)
}

func (s *Server) Delete(path string, handler Handler, resources ...*app.Resource) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}
	handlers := TransformHandlers(r, []Handler{handler}, s.logger)
	s.App.Delete(path, handlers...).Name(name)
}

func (s *Server) Patch(path string, handler Handler, resources ...*app.Resource) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}
	handlers := TransformHandlers(r, []Handler{handler}, s.logger)
	s.App.Patch(path, handlers...).Name(name)
}

func (s *Server) Options(path string, handler Handler, resources ...*app.Resource) {
	var r *app.Resource
	name := ""
	if len(resources) > 0 {
		name = resources[0].Name()
		r = resources[0]
	}
	handlers := TransformHandlers(r, []Handler{handler}, s.logger)
	s.App.Options(path, handlers...).Name(name)
}
