package restresolver_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := restresolver.Config{
		AppName:     "TestApp",
		JSONEncoder: json.Marshal,
		Logger:      app.CreateMockLogger(true),
	}

	server := restresolver.New(config)
	assert.NotNil(t, server.App)
}

func TestServerUse(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Use(func(c *restresolver.Context) error {
		c.Header("X-Test", "test")
		return c.Next()
	})

	req := httptest.NewRequest("GET", "/user", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, "test", resp.Header.Get("X-Test"))

	server.Group("profile", &app.Resource{}, func(c *restresolver.Context) error {
		return c.JSON("profile")
	})

	req = httptest.NewRequest("GET", "/profile", nil)
	resp, err = server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `"profile"`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestServerStatic(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	prefix := "/static"
	root := t.TempDir()

	err := utils.WriteFile(root+"/index.html", "index")
	assert.NoError(t, err)

	config := restresolver.StaticConfig{
		Index:         "index.html",
		Browse:        true,
		MaxAge:        3600,
		Compress:      true,
		ByteRange:     true,
		Download:      true,
		CacheDuration: 86400,
	}

	server.Static(prefix, root, config)

	// Test if the static route is registered correctly
	req := httptest.NewRequest("GET", "/static/index.html", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `index`, utils.Must(utils.ReadCloserToString(resp.Body)))

	// Test if the static route with a different file is not found
	req = httptest.NewRequest("GET", "/static/other.html", nil)
	resp, err = server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestServerMethods(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	methodsMap := map[string]func(path string, handler restresolver.Handler, resources ...*app.Resource){
		"GET":     server.Get,
		"HEAD":    server.Head,
		"POST":    server.Post,
		"PUT":     server.Put,
		"DELETE":  server.Delete,
		"CONNECT": server.Connect,
		"OPTIONS": server.Options,
		"TRACE":   server.Trace,
		"PATCH":   server.Patch,
	}

	for method, methodFunc := range methodsMap {
		methodFunc("/test", func(c *restresolver.Context) error {
			return c.JSON(method)
		}, &app.Resource{})

		req := httptest.NewRequest(method, "/test", nil)
		resp, err := server.Test(req)
		assert.NoError(t, err)
		defer closeResponse(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestServerListen(t *testing.T) {
	config := restresolver.Config{
		Logger: app.CreateMockLogger(true),
	}
	server := restresolver.New(config)
	go func() {
		time.Sleep(10 * time.Millisecond)
		server2 := restresolver.New(config)
		err := server2.Listen(":8080")
		assert.Error(t, err)
		assert.NoError(t, server.App.Shutdown())
	}()
	err := server.Listen(":8080")
	assert.NoError(t, err)
}
