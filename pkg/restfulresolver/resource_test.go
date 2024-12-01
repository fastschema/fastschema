package restfulresolver_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type testInput struct {
	Name string `json:"name"`
}

func TestNewRestResolver(t *testing.T) {
	resourceManager := fs.NewResourcesManager()
	resourceManager.Middlewares = []fs.Middleware{
		func(c fs.Context) error {
			c.Local("test", "test")
			return c.Next()
		},
	}

	staticDir := t.TempDir()
	utils.WriteFile(staticDir+"/test.txt", "test")
	staticFSs := []*fs.StaticFs{{
		BasePath: "/static",
		Root:     http.Dir(staticDir),
	}}

	resourceManager.Group("user", &fs.Meta{Prefix: "/userprefix"}).
		Add(fs.NewResource("profile", func(c fs.Context, input *testInput) (map[string]any, error) {
			return map[string]any{
				"input": input,
				"test":  c.Local("test"),
			}, nil
		}, &fs.Meta{Post: "/profile"})).
		Add(fs.NewResource("profileerror", func(c fs.Context, input *testInput) (map[string]any, error) {
			return nil, errors.BadRequest("test error")
		}, &fs.Meta{Post: "/profileerror"}))

	resourceManager.Add(fs.NewResource("bytes", func(c fs.Context, _ any) (any, error) {
		return []byte(`{"data": "test"}`), nil
	}))

	resourceManager.Add(fs.NewResource("html", func(c fs.Context, _ any) (any, error) {
		header := make(http.Header)
		header.Set("Content-Type", "text/html")

		return &fs.HTTPResponse{
			Header: header,
			Body:   []byte(`<body>test</body>`),
		}, nil
	}))

	resourceManager.Add(fs.NewResource("file", func(c fs.Context, _ any) (any, error) {
		header := make(http.Header)
		header.Set("Content-Type", "text/html")

		return &fs.HTTPResponse{
			Header: header,
			File:   staticDir + "/test.txt",
		}, nil
	}))

	resourceManager.Add(fs.NewResource("buffer", func(c fs.Context, _ any) (any, error) {
		header := make(http.Header)
		header.Set("Content-Type", "text/html")

		return &fs.HTTPResponse{
			Header: header,
			Stream: &bytes.Buffer{},
		}, nil
	}))

	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resourceManager,
		Logger:          logger.CreateMockLogger(true),
		StaticFSs:       staticFSs,
	})

	handlerFunc, err := restResolver.HTTPAdaptor()
	assert.NoError(t, err)
	assert.NotNil(t, handlerFunc)
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/static/test.txt", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `test`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req2 := httptest.NewRequest("POST", "/userprefix/profile", bytes.NewReader([]byte(`{"name": "test"}`)))
	req2.Header.Set("Content-Type", "application/json")
	resp, err = restResolver.Server().App.Test(req2)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `{"data":{"input":{"name":"test"},"test":"test"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req3 := httptest.NewRequest("POST", "/userprefix/profileerror", bytes.NewReader([]byte(`{"name": "test"}`)))
	req3.Header.Set("Content-Type", "application/json")
	resp, err = restResolver.Server().App.Test(req3)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"400","message":"test error"},"data":null}`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req4 := httptest.NewRequest("GET", "/bytes", nil)
	resp, err = restResolver.Server().App.Test(req4)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `{"data": "test"}`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req5 := httptest.NewRequest("GET", "/html", nil)
	resp, err = restResolver.Server().App.Test(req5)
	respContentType := resp.Header.Get("Content-Type")
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, "text/html", respContentType)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `<body>test</body>`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req6 := httptest.NewRequest("GET", "/file", nil)
	resp, err = restResolver.Server().App.Test(req6)
	respContentType = resp.Header.Get("Content-Type")
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, "text/html", respContentType)
	assert.Equal(t, 200, resp.StatusCode)

	req7 := httptest.NewRequest("GET", "/buffer", nil)
	resp, err = restResolver.Server().App.Test(req7)
	respContentType = resp.Header.Get("Content-Type")
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, "text/html", respContentType)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNewRestResolverErrorMiddleware(t *testing.T) {
	resourceManager := fs.NewResourcesManager()
	resourceManager.Middlewares = []fs.Middleware{
		func(c fs.Context) error {
			return errors.BadGateway("test bad gateway")
		},
	}

	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resourceManager,
		Logger:          logger.CreateMockLogger(true),
	})
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 502, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"502","message":"test bad gateway"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))

	resourceManager = fs.NewResourcesManager()
	resourceManager.Middlewares = []fs.Middleware{
		func(c fs.Context) error {
			return fiber.ErrBadRequest
		},
	}

	restResolver = restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resourceManager,
		Logger:          logger.CreateMockLogger(true),
	})
	assert.NotNil(t, restResolver.Server())

	req2 := httptest.NewRequest("GET", "/test", nil)
	resp, err = restResolver.Server().App.Test(req2)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestNewRestResolverWithBeforeResolverHookError(t *testing.T) {
	resourceManager := fs.NewResourcesManager()
	resourceManager.Add(fs.NewResource("test", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	}))
	resourceManager.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{
				func(c fs.Context) error {
					return errors.BadRequest("test before hook error")
				},
			},
		}
	}

	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resourceManager,
		Logger:          logger.CreateMockLogger(true),
	})
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"400","message":"test before hook error"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestNewRestResolverWithAfterResolverHookError(t *testing.T) {
	resourceManager := fs.NewResourcesManager()
	resourceManager.Add(fs.NewResource("test", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	}))
	resourceManager.Hooks = func() *fs.Hooks {
		return &fs.Hooks{
			PostResolve: []fs.Middleware{
				func(c fs.Context) error {
					return errors.BadRequest("test after hook error")
				},
			},
		}
	}

	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resourceManager,
		Logger:          logger.CreateMockLogger(true),
	})
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"400","message":"test after hook error"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestNewRestResolverStart(t *testing.T) {
	resourceManager := fs.NewResourcesManager()
	resourceManager.Add(fs.NewResource("test", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	}))

	restResolver := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resourceManager,
		Logger:          logger.CreateMockLogger(true),
	})

	go func() {
		time.Sleep(10 * time.Millisecond)
		server2 := restfulresolver.New(restfulresolver.Config{})
		err := server2.Listen(":8080")
		assert.Error(t, err)
		assert.NoError(t, restResolver.Shutdown())
	}()

	err := restResolver.Start(":8080")
	assert.NoError(t, err)
}
