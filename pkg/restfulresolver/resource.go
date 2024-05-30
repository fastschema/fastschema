package restfulresolver

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type RestfulResolver struct {
	resourceManager *fs.ResourcesManager
	staticFSs       []*fs.StaticFs
	server          *Server
}

func NewRestfulResolver(
	resourceManager *fs.ResourcesManager,
	logger logger.Logger,
	staticFSs ...*fs.StaticFs,
) *RestfulResolver {
	rs := &RestfulResolver{
		resourceManager: resourceManager,
		staticFSs:       staticFSs,
	}

	return rs.init(logger)
}

func (r *RestfulResolver) init(logger logger.Logger) *RestfulResolver {
	middlewares := []Handler{
		MiddlewareCors,
		MiddlewareRecover,
		MiddlewareRequestID,
		CreateMiddlewareRequestLog(r.staticFSs),
	}
	r.server = New(Config{
		AppName: "fastschema",
		Logger:  logger,
	})

	for _, middleware := range r.resourceManager.Middlewares {
		middlewares = append(middlewares, func(c *Context) error {
			if err := middleware(c); err != nil {
				fiberError, ok := err.(*fiber.Error)
				if ok {
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

	r.server.Use(middlewares...)

	for _, staticResource := range r.staticFSs {
		r.server.App.Use(staticResource.BasePath, filesystem.New(filesystem.Config{
			Root:       staticResource.Root,
			PathPrefix: staticResource.PathPrefix,
		}))
	}

	manager := r.server.Group(r.resourceManager.Name(), nil)

	var getHooks = func() *fs.Hooks {
		return &fs.Hooks{}
	}

	if r.resourceManager.Hooks != nil {
		getHooks = r.resourceManager.Hooks
	}

	registerResourceRoutes(r.resourceManager.Resources(), manager, getHooks)

	return r
}

func (r *RestfulResolver) Server() *Server {
	return r.server
}

func (r *RestfulResolver) Start(address string) error {
	return r.server.Listen(address)
}

func (r *RestfulResolver) Shutdown() error {
	return r.server.App.Shutdown()
}

type MethodData struct {
	Path    string
	Handler func(path string, handler Handler, resources ...*fs.Resource)
}

func registerResourceRoutes(
	resources []*fs.Resource,
	router *Router,
	getHooks func() *fs.Hooks,
) {
	for _, r := range resources {
		if r.IsGroup() {
			groupPrefix := r.Name()
			if r.Meta() != nil && r.Meta().Prefix != "" {
				groupPrefix = r.Meta().Prefix
			}

			group := router.Group(groupPrefix, r)
			registerResourceRoutes(r.Resources(), group, getHooks)

			continue
		}

		meta := r.Meta()
		path := r.Name()
		handler := router.Get

		if meta != nil {
			methods := []MethodData{{
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
			}}

			for _, method := range methods {
				if method.Path != "" {
					handler = method.Handler
					path = method.Path
				}
			}
		}

		hooks := getHooks()

		func(r *fs.Resource) {
			handler(path, func(c *Context) error {
				for _, hook := range hooks.PreResolve {
					if err := hook(c); err != nil {
						result := fs.NewResult(nil, err)
						if result.Error != nil && result.Error.Status != 0 {
							c.Status(result.Error.Status)
						}

						return c.JSON(result)
					}
				}

				result := fs.NewResult(r.Handler()(c))
				if result.Error != nil && result.Error.Status != 0 {
					c.Status(result.Error.Status)
				}

				c.Result(result)

				for _, hook := range hooks.PostResolve {
					if err := hook(c); err != nil {
						result := fs.NewResult(nil, err)

						if result.Error != nil && result.Error.Status != 0 {
							c.Status(result.Error.Status)
						}

						return c.JSON(result)
					}
				}

				// Send raw bytes
				bytes, ok := result.Data.([]byte)
				if ok {
					return c.Send(bytes)
				}

				// Send http response
				httpResponse, ok := result.Data.(*fs.HTTPResponse)
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

					return c.Status(status).Send(httpResponse.Body)
				}

				return c.JSON(result)
			}, r)
		}(r)
	}
}
