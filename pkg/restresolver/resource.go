package restresolver

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type StaticPaths struct {
	BasePath string
	Root     string
}

type RestSolver struct {
	resourceManager *app.ResourcesManager
	statics         []*StaticPaths
}

func NewRestResolver(resourceManager *app.ResourcesManager, statics []*StaticPaths) *RestSolver {
	return &RestSolver{
		resourceManager: resourceManager,
		statics:         statics,
	}
}

func (r RestSolver) Resource(routeName string) *app.Resource {
	return r.resourceManager.Find(routeName)
}

func (r *RestSolver) Start(address string, logger app.Logger) error {
	middlewares := []Handler{
		MiddlewareCors,
		MiddlewareRecover,
		MiddlewareRequestID,
		MiddlewareRequestLog,
	}
	s := New(Config{
		AppName: "fastschema",
		Logger:  logger,
	})

	for _, staticResource := range r.resourceManager.StaticResources {
		s.App.Use(staticResource.BasePath, filesystem.New(filesystem.Config{
			Root:       staticResource.Root,
			PathPrefix: staticResource.PathPrefix,
			Browse:     true,
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

	for _, static := range r.statics {
		s.Static(static.BasePath, static.Root)
	}

	s.Use(middlewares...)
	api := s.Group(r.resourceManager.Name(), nil)

	if err := registerResourceRoutes(
		r.resourceManager.Resources(),
		api,
		r.resourceManager.BeforeResolveHooks,
		r.resourceManager.AfterResolveHooks,
	); err != nil {
		panic(err)
	}

	s.Listen(address)

	return nil
}

func registerResourceRoutes(
	resources []*app.Resource,
	router *Router,
	beforeHandlerHooks []app.Middleware,
	afterHandlerHooks []app.Middleware,
) error {
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
			if err := registerResourceRoutes(
				r.Resources(),
				group,
				beforeHandlerHooks,
				afterHandlerHooks,
			); err != nil {
				return err
			}

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

	return nil
}
