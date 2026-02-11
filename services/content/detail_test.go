package contentservice_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestContentServiceDetail(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("GET", "/content/test/1", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid id (for uint64 schema)
	req = httptest.NewRequest("GET", "/content/blog/invalid", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `invalid`)

	blogModel := utils.Must(cs.DB().Model("blog"))
	blogID := utils.Must(blogModel.CreateFromJSON(context.Background(), `{"name": "test blog"}`))

	// Case 3: blog entity not found (non-existent numeric ID)
	req = httptest.NewRequest("GET", "/content/blog/999999", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `no entities found`)

	// Case 4: detail success
	req = httptest.NewRequest("GET", fmt.Sprintf("/content/blog/%v?select=name", blogID), nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"name":"test blog"`)

	userModel := utils.Must(cs.DB().Model("user"))
	userID := utils.Must(userModel.CreateFromJSON(context.Background(), `{"username": "testuser", "password": "123456", "provider": "local"}`))

	// Case 5: detail user entity should not have password field
	req = httptest.NewRequest("GET", fmt.Sprintf("/content/user/%v", userID), nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotContains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"password"`)
}
