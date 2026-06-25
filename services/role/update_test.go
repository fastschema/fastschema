package roleservice_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceUpdate(t *testing.T) {
	testApp := createTestApp()
	userRoleID := testApp.roleIDMap[fs.RoleUser.Name]

	// Case 1: Invalid Payload
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/role/%v", userRoleID), nil)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 2: Invalid ID
	req = httptest.NewRequest("PUT", "/api/role/00000000-0000-0000-0000-000000009999", bytes.NewReader([]byte(`{"name": "user role"}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)

	// Case 3: Attempt to rename a system role — must be rejected (system roles are identified by name)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/role/%v", userRoleID), bytes.NewReader([]byte(`{"name": "user role"}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, "Can't rename default roles")

	// Case 4: Role rule compile error (name unchanged — same as existing "User")
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/role/%v", userRoleID), bytes.NewReader([]byte(`{
		"name": "User",
		"rule": "invalid rule"
	}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)

	// Case 5: Permission compile error (name unchanged — same as existing "User")
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/role/%v", userRoleID), bytes.NewReader([]byte(`{
		"name": "User",
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

	// Case 6: Valid Payload, update role and permissions (name unchanged)
	// Current permissions: content.blog.list=allow, content.blog.detail=deny, content.blog.meta=notset
	// This test perform:
	// 	- Remove content.blog.detail
	// 	- Add content.blog.meta, content.blog.view
	//  - Update content.blog.list rule to $context.User().ID < 2
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/role/%v", userRoleID), bytes.NewReader([]byte(`{
		"name": "User",
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
			Where(db.EQ("id", userRoleID)).
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
