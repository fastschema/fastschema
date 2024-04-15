package roleservice_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceUpdate(t *testing.T) {
	testApp := createRoleTest()
	// Case 1: Invalid Payload
	req := httptest.NewRequest("PUT", "/role/2", nil)
	req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 2: Invalid ID
	req = httptest.NewRequest("PUT", "/role/9999", bytes.NewReader([]byte(`{"name": "user role"}`)))
	req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)

	// Case 3: Valid Payload, update role only
	req = httptest.NewRequest("PUT", "/role/2", bytes.NewReader([]byte(`{"name": "user role"}`)))
	req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"user role"`)

	// Case 4: Valid Payload, update role and permissions
	// Current permissions: content.list=allow, content.detail=deny, content.meta=notset
	// This test perform:
	// 	- Remove content.detail
	// 	- Add content.meta, content.view
	req = httptest.NewRequest("PUT", "/role/2", bytes.NewReader([]byte(`{
		"name": "user role",
		"permissions": [
			"content.list",
			"content.meta",
			"content.view"
		]
	}`)))
	req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	userRole := app.EntityToRole(utils.Must(testApp.roleModel.Query(app.EQ("id", 2)).
		Select("name", "permissions.resource", "permissions.value").
		First()))

	permissions := utils.Map(userRole.Permissions, func(p *app.Permission) string {
		return p.Resource
	})

	assert.Len(t, userRole.Permissions, 3)
	assert.Contains(t, permissions, "content.list")
	assert.Contains(t, permissions, "content.meta")
	assert.Contains(t, permissions, "content.view")
}
