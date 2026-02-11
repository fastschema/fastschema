package roleservice_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceDetail(t *testing.T) {
	testApp := createTestApp()
	userRoleID := testApp.roleIDMap[fs.RoleUser.Name]

	// Case 1: Invalid ID
	req := httptest.NewRequest("GET", "/api/role/00000000-0000-0000-0000-000000009999", nil)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)

	// Case 2: Success
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/role/%v", userRoleID), nil)
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"name":"User"`)
}
