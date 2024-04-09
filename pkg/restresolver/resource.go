package restresolver

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type RestSolver struct {
	resourceManager *app.ResourcesManager
	staticFSs       []*app.StaticFs
	server          *Server
}

func NewRestResolver(
	resourceManager *app.ResourcesManager,
	staticFSs ...*app.StaticFs,
) *RestSolver {
	return &RestSolver{
		resourceManager: resourceManager,
		staticFSs:       staticFSs,
	}
}

func (r *RestSolver) Init(logger app.Logger) *RestSolver {
	middlewares := []Handler{
		MiddlewareCors,
		MiddlewareRecover,
		MiddlewareRequestID,
		MiddlewareRequestLog,
	}
	r.server = New(Config{
		AppName: "fastschema",
		Logger:  logger,
	})

	for _, staticResource := range r.staticFSs {
		r.server.App.Use(staticResource.BasePath, filesystem.New(filesystem.Config{
			Root:       staticResource.Root,
			PathPrefix: staticResource.PathPrefix,
		}))
	}

	for _, middleware := range r.resourceManager.Middlewares {
		middlewares = append(middlewares, func(c *Context) error {
			if err := middleware(c); err != nil {
				fiberError, ok := err.(*fiber.Error)
				if ok {
					err = errors.GetErrorByStatus(fiberError.Code, err)
				}

				result := app.NewResult(nil, err)

				if result.Error != nil && result.Error.Status != 0 {
					c.Status(result.Error.Status)
				}

				return c.JSON(result)
			}

			return nil
		})
	}

	r.server.Use(middlewares...)
	manager := r.server.Group(r.resourceManager.Name(), nil)

	registerResourceRoutes(
		r.resourceManager.Resources(),
		manager,
		r.resourceManager.BeforeResolveHooks,
		r.resourceManager.AfterResolveHooks,
	)

	return r
}

func (r *RestSolver) Server() *Server {
	return r.server
}

func (r *RestSolver) Start(address string, logger app.Logger) error {
	r.Init(logger)
	return r.server.Listen(address)
}

func (r *RestSolver) Shutdown() error {
	return r.server.App.Shutdown()
}

func registerResourceRoutes(
	resources []*app.Resource,
	router *Router,
	beforeHandlerHooks []app.Middleware,
	afterHandlerHooks []app.Middleware,
) {
	methodMapper := map[string]func(string, Handler, ...*app.Resource){
		app.GET:     router.Get,
		app.POST:    router.Post,
		app.PUT:     router.Put,
		app.DELETE:  router.Delete,
		app.PATCH:   router.Patch,
		app.OPTIONS: router.Options,
	}

	for _, r := range resources {
		if r.IsGroup() {
			group := router.Group(r.Name(), r)
			registerResourceRoutes(
				r.Resources(),
				group,
				beforeHandlerHooks,
				afterHandlerHooks,
			)

			continue
		}

		meta := r.Meta()
		path := r.Name()
		handler := methodMapper[app.GET]

		for matchedMethod, matchedHandler := range methodMapper {
			if _, ok := meta[matchedMethod]; ok {
				handler = matchedHandler
				path = meta[matchedMethod].(string)
			}
		}

		func(r *app.Resource) {
			handler(path, func(c *Context) error {
				for _, hook := range beforeHandlerHooks {
					if err := hook(c); err != nil {
						result := app.NewResult(nil, err)
						if result.Error != nil && result.Error.Status != 0 {
							c.Status(result.Error.Status)
						}

						return c.JSON(result)
					}
				}

				result := app.NewResult(r.Resolver()(c))
				if result.Error != nil && result.Error.Status != 0 {
					c.Status(result.Error.Status)
				}

				c.Result(result)

				for _, hook := range afterHandlerHooks {
					if err := hook(c); err != nil {
						result := app.NewResult(nil, err)
						if result.Error != nil && result.Error.Status != 0 {
							c.Status(result.Error.Status)
						}

						return c.JSON(result)
					}
				}

				return c.JSON(result)
			}, r)
		}(r)
	}
}
