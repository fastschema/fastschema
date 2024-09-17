package fastschema

import (
	"net/http"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	authservice "github.com/fastschema/fastschema/services/auth"
	contentservice "github.com/fastschema/fastschema/services/content"
	fileservice "github.com/fastschema/fastschema/services/file"
	realtimeservice "github.com/fastschema/fastschema/services/realtime"
	roleservice "github.com/fastschema/fastschema/services/role"
	schemaservice "github.com/fastschema/fastschema/services/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
)

var (
	providerNames = strings.Join(fs.AuthProviders(), ", ")
	providerArgs  = fs.Args{
		"provider": fs.Arg{
			Required:    true,
			Type:        fs.TypeString,
			Description: "The auth provider name. Available providers: " + providerNames,
			Example:     "google",
		}}
	schemaArgs = fs.Args{
		"schema": {
			Required:    true,
			Type:        fs.TypeString,
			Description: "The schema name",
		}}
)

func (a *App) createResources() error {
	ms := fileservice.New(a)
	rs := roleservice.New(a)
	ss := schemaservice.New(a)
	cs := contentservice.New(a)
	ts := toolservice.New(a)
	as := authservice.New(a)
	rt := realtimeservice.New(a)

	a.config.Hooks.DBHooks.PostDBQuery = append(
		a.config.Hooks.DBHooks.PostDBQuery,
		ms.FileListHook,
	)
	a.config.Hooks.DBHooks.PostDBCreate = append(
		a.config.Hooks.DBHooks.PostDBCreate,
		rt.ContentCreateHook,
	)
	a.config.Hooks.DBHooks.PostDBUpdate = append(
		a.config.Hooks.DBHooks.PostDBUpdate,
		rt.ContentUpdateHook,
	)
	a.config.Hooks.DBHooks.PostDBDelete = append(
		a.config.Hooks.DBHooks.PostDBDelete,
		rt.ContentDeleteHook,
	)
	a.config.Hooks.PreResolve = append(
		a.config.Hooks.PreResolve,
		as.Authorize,
	)

	a.resources = fs.NewResourcesManager()
	a.resources.Middlewares = append(
		a.resources.Middlewares,
		as.ParseUser,
	)
	a.resources.Hooks = func() *fs.Hooks {
		return a.config.Hooks
	}

	a.api = a.resources.Group("api", &fs.Meta{Prefix: a.config.APIBaseName})

	if len(a.authProviders) > 0 {
		a.api.Group("auth").
			Add(fs.Get("me", as.Me, &fs.Meta{Public: true})).
			Group("provider", &fs.Meta{
				Prefix: "/:provider",
				Args:   providerArgs,
			}).
			Add(
				fs.NewResource("login", as.Login, createAuthMeta("/login")),
				fs.NewResource("callback", as.Callback, createAuthMeta("/register")),
			)
	}

	a.api.Group("schema").
		Add(fs.NewResource("list", ss.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("create", ss.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("detail", ss.Detail, &fs.Meta{
			Get:  "/:name",
			Args: fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("update", ss.Update, &fs.Meta{
			Put:  "/:name",
			Args: fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("delete", ss.Delete, &fs.Meta{
			Delete: "/:name",
			Args:   fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("import", ss.Import, &fs.Meta{Post: "/import"})).
		Add(fs.NewResource("export", ss.Export, &fs.Meta{Post: "/export"}))

	a.api.Group("content", &fs.Meta{Prefix: "/content/:schema", Args: schemaArgs}).
		Add(fs.NewResource("list", cs.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("detail", cs.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("create", cs.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("bulk-update", cs.BulkUpdate, &fs.Meta{
			Put: "/update",
		})).
		Add(fs.NewResource("update", cs.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("bulk-delete", cs.BulkDelete, &fs.Meta{
			Delete: "/delete",
		})).
		Add(fs.NewResource("delete", cs.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		}))

	a.api.Group("role").
		Add(fs.NewResource("list", rs.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("resources", rs.ResourcesList, &fs.Meta{
			Get: "/resources",
		})).
		Add(fs.NewResource("detail", rs.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("create", rs.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("update", rs.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("delete", rs.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		}))

	a.api.Group("file").
		Add(fs.NewResource("upload", ms.Upload, &fs.Meta{Post: "/upload"})).
		Add(fs.NewResource("delete", ms.Delete, &fs.Meta{Delete: "/"}))

	a.api.Group("tool").
		Add(fs.NewResource("stats", ts.Stats, &fs.Meta{Get: "/stats"}))

	a.api.
		Group("realtime").
		Add(fs.NewResource("content", rt.Content, &fs.Meta{WS: "/content"}))

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

	return nil
}

func createArg(t fs.ArgType, desc string) fs.Arg {
	return fs.Arg{Type: t, Required: true, Description: desc}
}

func createAuthMeta(path string) *fs.Meta {
	return &fs.Meta{
		Public: true,
		Post:   path,
		Get:    path,
	}
}
