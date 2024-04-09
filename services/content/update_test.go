package contentservice_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/app"
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
	blogID := utils.Must(blogModel.CreateFromJSON(`{"name": "test blog"}`))

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
	userID := utils.Must(userModel.CreateFromJSON(`{"username": "testuser", "password": "testpassword"}`))
	user := utils.Must(userModel.Query(app.EQ("id", userID)).First())

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
	userUpdated := utils.Must(userModel.Query(app.EQ("id", userID)).First())
	assert.NotEqual(t, user.GetString("password"), userUpdated.GetString("password"))
}
