package fastschema

import (
	"net/http"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	as "github.com/fastschema/fastschema/services/auth"
	cs "github.com/fastschema/fastschema/services/content"
	ms "github.com/fastschema/fastschema/services/file"
	realtime "github.com/fastschema/fastschema/services/realtime"
	rs "github.com/fastschema/fastschema/services/role"
	ss "github.com/fastschema/fastschema/services/schema"
	ts "github.com/fastschema/fastschema/services/tool"
	us "github.com/fastschema/fastschema/services/user"
)

func (a *App) createResources() error {
	userService := us.New(a)
	roleService := rs.New(a)
	fileService := ms.New(a)
	schemaService := ss.New(a)
	contentService := cs.New(a)
	toolService := ts.New(a)
	authService := as.New(a)
	realtimeService := realtime.New(a)

	a.hooks.DBHooks.PostDBGet = append(a.hooks.DBHooks.PostDBGet, fileService.FileListHook)
	a.hooks.DBHooks.PostDBCreate = append(a.hooks.DBHooks.PostDBCreate, realtimeService.ContentCreateHook)
	a.hooks.DBHooks.PostDBUpdate = append(a.hooks.DBHooks.PostDBUpdate, realtimeService.ContentUpdateHook)
	a.hooks.DBHooks.PostDBDelete = append(a.hooks.DBHooks.PostDBDelete, realtimeService.ContentDeleteHook)
	a.hooks.PreResolve = append(a.hooks.PreResolve, authService.Authorize)

	a.resources = fs.NewResourcesManager()
	a.resources.Middlewares = append(a.resources.Middlewares, authService.ParseUser)
	a.resources.Hooks = func() *fs.Hooks {
		return a.hooks
	}

	a.api = a.resources.Group("api", &fs.Meta{Prefix: a.config.APIBaseName})
	a.api.Group("user").
		Add(fs.NewResource("logout", userService.Logout, &fs.Meta{
			Post:   "/logout",
			Public: true,
		})).
		Add(fs.NewResource("me", userService.Me, &fs.Meta{Public: true})).
		Add(fs.NewResource("login", userService.Login, &fs.Meta{
			Post:   "/login",
			Public: true,
		}))

	if len(a.authProviders) > 0 {
		a.api.Group("auth", &fs.Meta{
			Prefix: "/auth/:provider",
			Args: fs.Args{
				"provider": {
					Required:    true,
					Type:        fs.TypeString,
					Description: "The auth provider name. Available providers: " + strings.Join(fs.AuthProviders(), ", "),
					Example:     "google",
				},
			},
		}).Add(
			fs.Get("login", authService.Login, &fs.Meta{Public: true}),
			fs.Get("callback", authService.Callback, &fs.Meta{Public: true}),
		)
	}

	a.api.Group("schema").
		Add(fs.NewResource("list", schemaService.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("create", schemaService.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("detail", schemaService.Detail, &fs.Meta{
			Get:  "/:name",
			Args: fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("update", schemaService.Update, &fs.Meta{
			Put:  "/:name",
			Args: fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("delete", schemaService.Delete, &fs.Meta{
			Delete: "/:name",
			Args:   fs.Args{"name": createArg(fs.TypeString, "The schema name")},
		}))

	a.api.Group("content", &fs.Meta{
		Prefix: "/content/:schema",
		Args: fs.Args{
			"schema": {
				Required:    true,
				Type:        fs.TypeString,
				Description: "The schema name",
			},
		},
	}).
		Add(fs.NewResource("list", contentService.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("detail", contentService.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("create", contentService.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("bulk-update", contentService.BulkUpdate, &fs.Meta{
			Put: "/update",
		})).
		Add(fs.NewResource("update", contentService.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("delete", contentService.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		}))

	a.api.Group("role").
		Add(fs.NewResource("list", roleService.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("resources", roleService.ResourcesList, &fs.Meta{
			Get: "/resources",
		})).
		Add(fs.NewResource("detail", roleService.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("create", roleService.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("update", roleService.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("delete", roleService.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The role ID")},
		}))

	a.api.Group("file").
		Add(fs.NewResource("upload", fileService.Upload, &fs.Meta{Post: "/upload"})).
		Add(fs.NewResource("delete", fileService.Delete, &fs.Meta{Delete: "/"}))

	a.api.Group("tool").
		Add(fs.NewResource("stats", toolService.Stats, &fs.Meta{
			Get:    "/stats",
			Public: true,
		}))

	a.api.
		Group("realtime").
		Add(fs.NewResource("content", realtimeService.Content, &fs.Meta{WS: "/content"}))

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
