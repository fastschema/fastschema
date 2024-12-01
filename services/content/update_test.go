package contentservice_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestContentServiceUpdate(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("PUT", "/content/test/1", bytes.NewReader([]byte(`{"name": "test blog"}`)))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid json
	req = httptest.NewRequest("PUT", "/content/blog/1", bytes.NewReader([]byte(`{"name": "invalid json"`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	blogModel := utils.Must(cs.DB().Model("blog"))
	blogID := utils.Must(blogModel.CreateFromJSON(context.Background(), `{"name": "test blog"}`))

	// Case 3: blog entity has invalid field
	req = httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/content/blog/%d", blogID),
		bytes.NewReader([]byte(`{"invalid": "test blog"}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(
		t,
		utils.Must(utils.ReadCloserToString(resp.Body)),
		`field $set.invalid error: column blog.invalid not found`,
	)

	// Case 4: update success
	req = httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/content/blog/%d", blogID),
		bytes.NewReader([]byte(`{"name": "updated name"}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(
		t,
		utils.Must(utils.ReadCloserToString(resp.Body)),
		`"name":"updated name"`,
	)

	userModel := utils.Must(cs.DB().Model("user"))
	userID := utils.Must(userModel.CreateFromJSON(context.Background(), `{"username": "testuser", "password": "testpassword", "provider": "local"}`))
	user := utils.Must(userModel.Query(db.EQ("id", userID)).First(context.Background()))

	// Case 5: update user without password
	req = httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/content/user/%d", userID),
		bytes.NewReader([]byte(`{"username": "updatedusername"}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"username":"updatedusername"`)

	// Case 6: update user with password
	req = httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/content/user/%d", userID),
		bytes.NewReader([]byte(`{"username": "updated", "password": "updatedpassword"}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"username":"updated"`)
	ctx := context.WithValue(context.Background(), "keeppassword", "true")
	userUpdated := utils.Must(userModel.Query(db.EQ("id", userID)).First(ctx))
	assert.NotEqual(t, user.GetString("password"), userUpdated.GetString("password"))
}

func TestContentServiceBulkUpdate(t *testing.T) {
	cs, server := createContentService(t)
	// Case 1: schema not found
	req := httptest.NewRequest("PUT", "/content/test/update", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// Case 2: invalid predicate
	req = httptest.NewRequest("PUT", "/content/blog/update?filter=invalid", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 3: invalid payload
	req = httptest.NewRequest("PUT", `/content/blog/update?filter={"name":{"$like":"%blog2%"}}`, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// create 10 blog posts
	blogModel := utils.Must(cs.DB().Model("blog"))
	for i := 0; i < 10; i++ {
		utils.Must(blogModel.CreateFromJSON(context.Background(), fmt.Sprintf(`{"name": "blog%d"}`, i+1)))
	}

	// Case 4: bulk update success
	filter := url.QueryEscape(`{"name":{"$like":"%blog2%"}}`)
	req = httptest.NewRequest(
		"PUT",
		"/content/blog/update?filter="+filter,
		bytes.NewReader([]byte(`{"name": "updated name"}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(
		t,
		utils.Must(utils.ReadCloserToString(resp.Body)),
		`"data":1`,
	)

	// Case 5: bulk update success with multiple predicates
	filter2 := url.QueryEscape(`{"name":{"$like":"%blog%"}}`)
	req = httptest.NewRequest(
		"PUT",
		"/content/blog/update?filter="+filter2,
		bytes.NewReader([]byte(`{"name": "updated name"}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(
		t,
		utils.Must(utils.ReadCloserToString(resp.Body)),
		`"data":9`,
	)

	// Case 6: bulk update with empty body
	filter3 := url.QueryEscape(`{"name":{"$like":"%updated name%"}}`)
	req = httptest.NewRequest(
		"PUT",
		"/content/blog/update?filter="+filter3,
		bytes.NewReader([]byte(`{}`)),
	)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(
		t,
		utils.Must(utils.ReadCloserToString(resp.Body)),
		`"data":0`,
	)
}
