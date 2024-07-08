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

func TestContentServiceBulkDelete(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("POST", "/content/test/bulk-delete", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid predicate
	req = httptest.NewRequest("POST", "/content/blog/bulk-delete?filter=invalid", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 3: filter not found
	filter := url.QueryEscape(`{"name":{"$like":"%test%"}}`)
	req = httptest.NewRequest("POST", "/content/blog/bulk-delete?filter="+filter, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)

	blogModel := utils.Must(cs.DB().Model("blog"))
	utils.Must(blogModel.CreateFromJSON(context.Background(), `{"name": "test blog"}`))
	userModel := utils.Must(cs.DB().Model("user"))
	utils.Must(userModel.CreateFromJSON(context.Background(), `{"username": "admin"}`))

	// Case 4: delete fail with root user
	filterUser := url.QueryEscape(`{"username":{"$like":"%admin%"}}`)
	req = httptest.NewRequest("POST", "/content/user/bulk-delete?filter="+filterUser, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"Cannot delete root user."`)

	// Case 5: delete success
	req = httptest.NewRequest("POST", "/content/blog/bulk-delete?filter="+filter, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestContentServiceBulkDeleteFail(t *testing.T) {
	cs, server := createContentService(t)

	blogModel := utils.Must(cs.DB().Model("blog"))
	blogID := utils.Must(blogModel.CreateFromJSON(context.Background(), `{"name": "test blog"}`))
	tagModel := utils.Must(cs.DB().Model("tag"))
	tagJSON := fmt.Sprintf(`{"name": "test tag", "blogs_id": %d}`, blogID)
	tagModel.CreateFromJSON(context.Background(), tagJSON)

	// Case 6: fail to delete
	filter := url.QueryEscape(`{"name":{"$like":"%test blog%"}}`)
	req := httptest.NewRequest("POST", "/content/blog/bulk-delete?filter="+filter, nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
}
