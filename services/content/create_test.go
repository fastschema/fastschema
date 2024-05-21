package contentservice_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestContentServiceCreateRequestError(t *testing.T) {
	_, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("POST", "/content/test", bytes.NewReader([]byte(`{"name": "test blog"}`)))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid json
	req = httptest.NewRequest("POST", "/content/blog", bytes.NewReader([]byte(`{"name": "invalid json"`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 3: blog entity has invalid field
	req = httptest.NewRequest("POST", "/content/blog", bytes.NewReader([]byte(`{"invalid": "test blog"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `column error: column blog.invalid not found`)
}

func TestContentServiceCreateNormal(t *testing.T) {
	cs, server := createContentService(t)
	req := httptest.NewRequest("POST", "/content/blog", bytes.NewReader([]byte(`{"name": "test blog"}`)))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	// check if blog is created
	blogModel := utils.Must(cs.DB().Model("blog"))
	blog := utils.Must(blogModel.Query(db.EQ("name", "test blog")).First(context.Background()))
	assert.Equal(t, "test blog", blog.GetString("name"))
	assert.Equal(t, uint64(1), blog.ID())
}

func TestContentServiceCreateUser(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: There is no password field
	req := httptest.NewRequest("POST", "/content/user", bytes.NewReader([]byte(`{"name": "test blog"}`)))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 2: Success
	req = httptest.NewRequest("POST", "/content/user", bytes.NewReader([]byte(`{"username": "testuser", "password": "testpassword"}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"id":1`)

	// check if user is created
	userModel := utils.Must(cs.DB().Model("user"))
	user := utils.Must(userModel.Query(db.EQ("username", "testuser")).First(context.Background()))
	assert.Equal(t, "testuser", user.GetString("username"))
	assert.NotEqual(t, "testpassword", user.GetString("password"))
}
