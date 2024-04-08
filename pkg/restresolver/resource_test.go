package restresolver_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type testInput struct {
	Name string `json:"name"`
}

func TestNewRestResolver(t *testing.T) {
	resourceManager := app.NewResourcesManager()
	resourceManager.Middlewares = []app.Middleware{
		func(c app.Context) error {
			c.Value("test", "test")
			return c.Next()
		},
	}

	staticDir := t.TempDir()
	utils.WriteFile(staticDir+"/test.txt", "test")
	staticFSs := []*app.StaticFs{{
		BasePath: "/static",
		Root:     http.Dir(staticDir),
	}}

	resourceManager.Group("user").
		Add(app.NewResource("profile", func(c app.Context, input *testInput) (map[string]any, error) {
			return map[string]any{
				"input": input,
				"test":  c.Value("test"),
			}, nil
		}, app.Meta{app.POST: "/profile"})).
		Add(app.NewResource("profileerror", func(c app.Context, input *testInput) (map[string]any, error) {
			return nil, errors.BadRequest("test error")
		}, app.Meta{app.POST: "/profileerror"}))

	restResolver := restresolver.NewRestResolver(resourceManager, staticFSs...)
	restResolver.Init(testutils.CreateLogger(true))
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/static/test.txt", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `test`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req2 := httptest.NewRequest("POST", "/user/profile", bytes.NewReader([]byte(`{"name": "test"}`)))
	req2.Header.Set("Content-Type", "application/json")
	resp, err = restResolver.Server().App.Test(req2)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `{"data":{"input":{"name":"test"},"test":"test"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))

	req3 := httptest.NewRequest("POST", "/user/profileerror", bytes.NewReader([]byte(`{"name": "test"}`)))
	req3.Header.Set("Content-Type", "application/json")
	resp, err = restResolver.Server().App.Test(req3)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"400","message":"test error"},"data":null}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestNewRestResolverErrorMiddleware(t *testing.T) {
	resourceManager := app.NewResourcesManager()
	resourceManager.Middlewares = []app.Middleware{
		func(c app.Context) error {
			return errors.BadGateway("test bad gateway")
		},
	}

	restResolver := restresolver.NewRestResolver(resourceManager, nil)
	restResolver.Init(testutils.CreateLogger(true))
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 502, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"502","message":"test bad gateway"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))

	resourceManager.Middlewares = []app.Middleware{
		func(c app.Context) error {
			return fiber.ErrBadRequest
		},
	}

	restResolver.Init(testutils.CreateLogger(true))
	assert.NotNil(t, restResolver.Server())

	req2 := httptest.NewRequest("GET", "/test", nil)
	resp, err = restResolver.Server().App.Test(req2)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestNewRestResolverWithBeforeResolverHookError(t *testing.T) {
	resourceManager := app.NewResourcesManager()
	resourceManager.Add(app.NewResource("test", func(c app.Context, _ *any) (any, error) {
		return nil, nil
	}))
	resourceManager.BeforeResolveHooks = []app.Middleware{
		func(c app.Context) error {
			return errors.BadRequest("test before hook error")
		},
	}

	restResolver := restresolver.NewRestResolver(resourceManager, nil)
	restResolver.Init(testutils.CreateLogger(true))
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"400","message":"test before hook error"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestNewRestResolverWithAfterResolverHookError(t *testing.T) {
	resourceManager := app.NewResourcesManager()
	resourceManager.Add(app.NewResource("test", func(c app.Context, _ *any) (any, error) {
		return nil, nil
	}))
	resourceManager.AfterResolveHooks = []app.Middleware{
		func(c app.Context) error {
			return errors.BadRequest("test after hook error")
		},
	}

	restResolver := restresolver.NewRestResolver(resourceManager, nil)
	restResolver.Init(testutils.CreateLogger(true))
	assert.NotNil(t, restResolver.Server())

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := restResolver.Server().App.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":{"code":"400","message":"test after hook error"}}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestNewRestResolverStart(t *testing.T) {
	resourceManager := app.NewResourcesManager()
	resourceManager.Add(app.NewResource("test", func(c app.Context, _ *any) (any, error) {
		return nil, nil
	}))

	restResolver := restresolver.NewRestResolver(resourceManager, nil)

	go func() {
		time.Sleep(10 * time.Millisecond)
		server2 := restresolver.New(restresolver.Config{})
		err := server2.Listen(":8080")
		assert.Error(t, err)
		assert.NoError(t, restResolver.Shutdown())
	}()

	err := restResolver.Start(":8080", testutils.CreateLogger(true))
	assert.NoError(t, err)
}
