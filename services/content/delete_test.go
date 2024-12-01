package contentservice_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestContentServiceDelete(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("DELETE", "/content/test/1", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid id
	req = httptest.NewRequest("DELETE", "/content/blog/2", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `no entities found`)

	blogModel := utils.Must(cs.DB().Model("blog"))
	blogID := utils.Must(blogModel.CreateFromJSON(context.Background(), `{"name": "test blog"}`))

	// Case 3: success
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/content/blog/%d", blogID), nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestContentServiceDeleteRootUser(t *testing.T) {
	_, server := createContentService(t)

	req := httptest.NewRequest("DELETE", "/content/user/1", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"Cannot delete root user."`)
}

func TestContentServiceBulkDelete(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("DELETE", "/content/test/delete", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid predicate
	req = httptest.NewRequest("DELETE", "/content/blog/delete?filter=invalid", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 3: filter not found
	filter := url.QueryEscape(`{"name":{"$like":"%test%"}}`)
	req = httptest.NewRequest("DELETE", "/content/blog/delete?filter="+filter, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	blogModel := utils.Must(cs.DB().Model("blog"))
	utils.Must(blogModel.CreateFromJSON(context.Background(), `{"name": "test blog"}`))
	userModel := utils.Must(cs.DB().Model("user"))
	utils.Must(userModel.CreateFromJSON(context.Background(), `{"username": "admin", "provider": "local"}`))

	// Case 4: delete fail with root user
	filterUser := url.QueryEscape(`{"username":{"$like":"%admin%"}}`)
	req = httptest.NewRequest("DELETE", "/content/user/delete?filter="+filterUser, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"Cannot delete root user."`)

	// Case 5: delete success
	req = httptest.NewRequest("DELETE", "/content/blog/delete?filter="+filter, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestContentServiceBulkDeleteFail(t *testing.T) {
	cs, server := createContentService(t)

	tagModel := utils.Must(cs.DB().Model("tag"))
	tagID := utils.Must(tagModel.CreateFromJSON(context.Background(), `{"name": "test tag"}`))
	blogModel := utils.Must(cs.DB().Model("blog"))
	blogJSON := fmt.Sprintf(`{"name": "test blog", "tags_id": %d}`, tagID)
	utils.Must(blogModel.CreateFromJSON(context.Background(), blogJSON))

	// Case 6: fail to delete
	filter := url.QueryEscape(`{"name":{"$like":"%test tag%"}}`)
	req := httptest.NewRequest("DELETE", "/content/tag/delete?filter="+filter, nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}
