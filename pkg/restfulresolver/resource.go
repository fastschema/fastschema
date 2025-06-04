package restfulresolver

import (
	"net/http"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type RestfulResolver struct {
	config *ResolverConfig
	server *Server
}

type ResolverConfig struct {
	ResourceManager *fs.ResourcesManager
	Logger          logger.Logger
	StaticFSs       []*fs.StaticFs
}

func NewRestfulResolver(config *ResolverConfig) *RestfulResolver {
	rs := &RestfulResolver{
		config: config,
	}

	return rs.init(config.Logger)
}

func (r *RestfulResolver) init(logger logger.Logger) *RestfulResolver {
	r.server = New(Config{
		AppName: "fastschema",
		Logger:  logger,
	})
	r.server.Use(append([]Handler{
		MiddlewareCors,
		MiddlewareRecover,
		MiddlewareRequestID,
		MiddlewareCookie,
		CreateMiddlewareRequestLog(r.config.StaticFSs),
	}, TransformMiddlewares(r.config.ResourceManager.Middlewares)...)...)

	// Static files
	for _, staticResource := range r.config.StaticFSs {
		r.server.App.Use(staticResource.BasePath, filesystem.New(filesystem.Config{
			Root:       staticResource.Root,
			PathPrefix: staticResource.PathPrefix,
			Next: func(c *fiber.Ctx) bool {
				// skip serving static files for root path
				return c.Path() == "/"
			},
		}))
	}

	// The public directory is served at root path
	// If the request path file is not found,
	// filesystem will set status to 404
	// We need to reset status to 200 because
	// we want to continue to resolve the request
	r.server.App.Use(func(c *fiber.Ctx) error {
		c.Status(fiber.StatusOK)
		return c.Next()
	})

	var getHooks = func() *fs.Hooks {
		return &fs.Hooks{}
	}

	if r.config.ResourceManager.Hooks != nil {
		getHooks = r.config.ResourceManager.Hooks
	}

	manager := r.server.Group(r.config.ResourceManager.Name(), nil)
	RegisterResourceRoutes(r.config.ResourceManager.Resources(), manager, getHooks)

	return r
}

func (r *RestfulResolver) Server() *Server {
	return r.server
}

func (r *RestfulResolver) HTTPAdaptor() (http.HandlerFunc, error) {
	return adaptor.FiberApp(r.Server().App), nil
}

func (r *RestfulResolver) Start(address string) error {
	return r.server.Listen(address)
}

func (r *RestfulResolver) Shutdown() error {
	return r.server.Shutdown()
}

func RegisterResourceRoutes(
	resources []*fs.Resource,
	router *Router,
	getHooks func() *fs.Hooks,
) {
	var hooks *fs.Hooks

	if getHooks != nil {
		hooks = getHooks()
	}

	for _, r := range resources {
		if r.IsGroup() {
			groupPrefix := r.Name()
			if r.Meta() != nil && r.Meta().Prefix != "" {
				groupPrefix = r.Meta().Prefix
			}

			group := router.Group(groupPrefix, r)
			RegisterResourceRoutes(r.Resources(), group, getHooks)
			continue
		}

		meta := r.Meta()
		httpHandlers := []MethodData{{
			Path:    r.Name(),
			Handler: router.Get,
		}}

		if meta != nil {
			metaHandlers := CreateHTTPHandlers(router, meta)
			if len(metaHandlers) > 0 {
				httpHandlers = metaHandlers
			}

			if meta.WS != "" {
				WSResourceHandler(r, hooks, router)
			}
		}

		// Register all available methods
		for _, methodHandler := range httpHandlers {
			httpResourceHandler(r, hooks, methodHandler)
		}
	}
}

func httpResourceHandler(r *fs.Resource, hooks *fs.Hooks, methodHandler MethodData) {
	methodHandler.Handler(methodHandler.Path, func(c *Context) error {
		for _, hook := range hooks.PreResolve {
			if err := hook(c); err != nil {
				if c.result = fs.NewResult(nil, err); c.result.Error.Status != 0 {
					c.Status(c.result.Error.Status)
				}

				return c.JSON(c.result)
			}
		}

		c.result = fs.NewResult(r.Handler()(c))
		for _, hook := range hooks.PostResolve {
			if err := hook(c); err != nil {
				c.result = fs.NewResult(nil, err)
				break
			}
		}

		if c.result.Error != nil {
			if c.result.Error.Status != 0 {
				c.Status(c.result.Error.Status)
			}
			return c.JSON(c.result)
		}

		// Send raw bytes
		bytes, ok := c.result.Data.([]byte)
		if ok {
			return c.Send(bytes)
		}

		// Send http response
		httpResponse, ok := c.result.Data.(*fs.HTTPResponse)
		if ok {
			status := httpResponse.StatusCode
			if status == 0 {
				status = fiber.StatusOK
			}

			if httpResponse.Header != nil {
				for key, values := range httpResponse.Header {
					for _, value := range values {
						c.Set(key, value)
					}
				}
			}

			if httpResponse.File != "" {
				return c.Ctx.SendFile(httpResponse.File)
			}

			if httpResponse.Stream != nil {
				return c.SendStream(httpResponse.Stream)
			}

			return c.Status(status).Send(httpResponse.Body)
		}

		return c.JSON(c.result)
	}, r)
}
