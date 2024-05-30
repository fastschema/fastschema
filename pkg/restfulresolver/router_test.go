package restfulresolver_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRouterUse(t *testing.T) {
	server := restfulresolver.New(restfulresolver.Config{})
	router := server.Group("user", nil)
	router.Use(func(c *restfulresolver.Context) error {
		c.Header("X-Test", "test")
		return c.Next()
	})

	req := httptest.NewRequest("GET", "/user/", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, "test", resp.Header.Get("X-Test"))

	router.Group("profile", &fs.Resource{}, func(c *restfulresolver.Context) error {
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
	server := restfulresolver.New(restfulresolver.Config{})
	router := server.Group("user", nil)
	methodsMap := map[string]func(path string, handler restfulresolver.Handler, resources ...*fs.Resource){
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
		methodFunc("/test", func(c *restfulresolver.Context) error {
			return c.JSON(method)
		}, &fs.Resource{})

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
