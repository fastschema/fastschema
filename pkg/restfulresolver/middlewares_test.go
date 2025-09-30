package restfulresolver_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewares(t *testing.T) {
	mockLogger := logger.CreateMockLogger(true)
	server := restfulresolver.New(restfulresolver.Config{
		Logger: mockLogger,
	})
	server.Use(
		restfulresolver.MiddlewareCookie,
		restfulresolver.MiddlewareRequestID,
		restfulresolver.CreateMiddlewareRequestLog([]*fs.StaticFs{}),
		restfulresolver.MiddlewareCors,
		restfulresolver.MiddlewareRecover,
	)
	server.Get("/test", func(c *restfulresolver.Context) error {
		return errors.New("test error")
	})
	server.Get("/panic", func(c *restfulresolver.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 500, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Request-Id"))
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Range", resp.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, 1, len(mockLogger.Messages))
	assert.Contains(t, mockLogger.Last().String(), "Request completed")

	req2 := httptest.NewRequest("OPTIONS", "/not-found", nil)
	resp, err = server.Test(req2)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	req3 := httptest.NewRequest("GET", "/panic", nil)
	resp, err = server.Test(req3)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, `{"error":"test panic"}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}
