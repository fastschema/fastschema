package roleservice_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceUpdate(t *testing.T) {
	testApp := createTestApp()
	// Case 1: Invalid Payload
	req := httptest.NewRequest("PUT", "/api/role/2", nil)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 2: Invalid ID
	req = httptest.NewRequest("PUT", "/api/role/9999", bytes.NewReader([]byte(`{"name": "user role"}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)

	// Case 3: Valid Payload, update role only
	req = httptest.NewRequest("PUT", "/api/role/2", bytes.NewReader([]byte(`{"name": "user role"}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"user role"`)

	// Case 4: Role rule compile error
	req = httptest.NewRequest("PUT", "/api/role/2", bytes.NewReader([]byte(`{
		"name": "user role",
		"rule": "invalid rule"
	}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)

	// Case 5: Permission compile error
	req = httptest.NewRequest("PUT", "/api/role/2", bytes.NewReader([]byte(`{
		"name": "user role",
		"permissions": [
			{
				"resource": "content.blog.list",
				"value": "invalid rule"
			}
		]
	}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)

	// Case 4: Valid Payload, update role and permissions
	// Current permissions: content.blog.list=allow, content.blog.detail=deny, content.blog.meta=notset
	// This test perform:
	// 	- Remove content.blog.detail
	// 	- Add content.blog.meta, content.blog.view
	//  - Update content.blog.list rule to $context.User().ID < 2
	req = httptest.NewRequest("PUT", "/api/role/2", bytes.NewReader([]byte(`{
		"name": "user role",
		"permissions": [
			{
				"resource": "content.blog.list",
				"value": "$context.User().ID < 2"
			},
			{
				"resource": "content.blog.meta",
				"value": "allow"
			},
			{
				"resource": "content.blog.view",
				"value": "allow"
			}
		]
	}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	userRole := utils.Must(
		db.Builder[*fs.Role](testApp.db).
			Where(db.EQ("id", 2)).
			Select("permissions").
			First(context.Background()),
	)

	permissions := utils.Map(userRole.Permissions, func(p *fs.Permission) string {
		return p.Resource
	})

	assert.Len(t, userRole.Permissions, 3)
	assert.Contains(t, permissions, "content.blog.list")
	assert.Contains(t, permissions, "content.blog.meta")
	assert.Contains(t, permissions, "content.blog.view")
}
