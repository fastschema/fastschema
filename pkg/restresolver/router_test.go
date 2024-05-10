package restresolver_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRouterUse(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	router := server.Group("user", nil)
	router.Use(func(c *restresolver.Context) error {
		c.Header("X-Test", "test")
		return c.Next()
	})

	req := httptest.NewRequest("GET", "/user/", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, "test", resp.Header.Get("X-Test"))

	router.Group("profile", &app.Resource{}, func(c *restresolver.Context) error {
		return c.JSON("profile")
	})

	req = httptest.NewRequest("GET", "/user/profile", nil)
	resp, err = server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `"profile"`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestRouterMethods(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	router := server.Group("user", nil)
	methodsMap := map[string]func(path string, handler restresolver.Handler, resources ...*app.Resource){
		"GET":     router.Get,
		"HEAD":    router.Head,
		"POST":    router.Post,
		"PUT":     router.Put,
		"DELETE":  router.Delete,
		"CONNECT": router.Connect,
		"OPTIONS": router.Options,
		"TRACE":   router.Trace,
		"PATCH":   router.Patch,
	}

	for method, methodFunc := range methodsMap {
		methodFunc("/test", func(c *restresolver.Context) error {
			return c.JSON(method)
		}, &app.Resource{})

		req := httptest.NewRequest(method, "/user/test", nil)
		resp, err := server.Test(req)
		assert.NoError(t, err)
		defer closeResponse(t, resp)
		assert.Equal(t, 200, resp.StatusCode)

		if method != "HEAD" {
			assert.Equal(t, `"`+method+`"`, utils.Must(utils.ReadCloserToString(resp.Body)))
		}
	}
}
