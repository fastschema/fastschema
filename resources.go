package fastschema

import (
	"net/http"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/fastschema/fastschema/services"
)

var ignoreContentSchemas = []string{
	"user",
	"role",
	"permission",
	"roles_users",
}

type AppConfig struct {
	Version   string           `json:"version"`
	Schemas   []*schema.Schema `json:"schemas"`
	Resources []*fs.Resource   `json:"resources"`
}

func (a *App) createServices() {
	a.services = services.New(a)
	realTimeService := a.services.Realtime()

	a.config.Hooks.DBHooks.PostDBQuery = append(
		a.config.Hooks.DBHooks.PostDBQuery,
		a.services.File().FileListHook,
	)
	a.config.Hooks.DBHooks.PostDBCreate = append(
		a.config.Hooks.DBHooks.PostDBCreate,
		realTimeService.ContentCreateHook,
	)
	a.config.Hooks.DBHooks.PostDBUpdate = append(
		a.config.Hooks.DBHooks.PostDBUpdate,
		realTimeService.ContentUpdateHook,
	)
	a.config.Hooks.DBHooks.PostDBDelete = append(
		a.config.Hooks.DBHooks.PostDBDelete,
		realTimeService.ContentDeleteHook,
	)
	a.config.Hooks.PreResolve = append(
		a.config.Hooks.PreResolve,
		a.services.Auth().Authorize,
	)
}

func (a *App) createResources() {
	a.resources = fs.NewResourcesManager()
	a.resources.Middlewares = append(
		a.resources.Middlewares,
		a.services.Auth().ParseUser,
	)
	a.resources.Hooks = func() *fs.Hooks {
		return a.config.Hooks
	}

	a.api = a.resources.Group("api", &fs.Meta{Prefix: a.config.APIBaseName})
	a.services.Auth().CreateResource(a.api, a.authProviders)
	a.services.Schema().CreateResource(a.api)
	a.services.Content().CreateResource(a.api)
	a.services.Realtime().CreateResource(a.api)
	a.services.Role().CreateResource(a.api)
	a.services.File().CreateResource(a.api)
	a.services.Tool().CreateResource(a.api)

	a.api.Add(fs.Get("config", func(c fs.Context, _ any) (*AppConfig, error) {
		schemas, err := a.services.Schema().List(c, nil)
		if err != nil {
			return nil, err
		}
		return &AppConfig{
			Version:   Version,
			Schemas:   schemas,
			Resources: a.ResourcesList(),
		}, nil
	}))

	a.resources.Group("docs").
		Add(fs.NewResource("spec", func(c fs.Context, _ any) (any, error) {
			return a.CreateOpenAPISpec()
		}, &fs.Meta{Get: "/openapi.json"})).
		Add(fs.NewResource("viewer", func(c fs.Context, _ any) (any, error) {
			header := make(http.Header)
			header.Set("Content-Type", "text/html")

			return &fs.HTTPResponse{
				StatusCode: http.StatusOK,
				Header:     header,
				Body: []byte(utils.CreateSwaggerUIPage(
					a.config.BaseURL + "/docs/openapi.json",
				)),
			}, nil
		}, &fs.Meta{Get: "/"}))
}

func (a *App) ResourcesList() []*fs.Resource {
	schemas := a.DB().SchemaBuilder().Schemas()
	resources := a.resources.Clone()
	apiResources := []*fs.Resource{}
	contentResource := &fs.Resource{}
	apiGroup := resources.Find("api")
	if apiGroup != nil {
		apiResources = apiGroup.Resources()
	}

	apiGroup.Print()
	for _, r := range apiResources {
		if r.Name() == "content" {
			contentResource = r
			apiGroup.Remove(r.Name())
		}
		if r.Name() == "realtime" {
			apiGroup.Remove(r.Name())
		}
	}

	contentGroup := apiGroup.Group("content")
	realtimeContentGroup := apiGroup.Group("realtime").Group("content")
	for _, schema := range schemas {
		if utils.Contains(ignoreContentSchemas, schema.Name) || schema.IsJunctionSchema {
			continue
		}

		schemaContentGroup := contentGroup.Group(schema.Name)
		for _, r := range contentResource.Resources() {
			schemaContentGroup.AddResource(r.Name(), nil, r.Meta().Clone())
		}

		realtimeSchemaGroup := realtimeContentGroup.Group(schema.Name)
		realtimeSchemaGroup.AddResource("*", nil, &fs.Meta{})
		realtimeSchemaGroup.AddResource("create", nil, &fs.Meta{})
		realtimeSchemaGroup.AddResource("update", nil, &fs.Meta{})
		realtimeSchemaGroup.AddResource("delete", nil, &fs.Meta{})
	}

	return resources.Resources()
}
