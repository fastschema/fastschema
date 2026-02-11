package roleservice_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceDelete(t *testing.T) {
	testApp := createTestApp()
	// Case 1: Invalid ID
	req := httptest.NewRequest("DELETE", "/api/role/00000000-0000-0000-0000-000000009999", nil)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)

	// Case 2: Delete default role - use actual role ID from database
	roles := utils.Must(testApp.roleModel.Query().Get(context.Background()))
	assert.True(t, len(roles) > 0, "Should have at least one role")
	defaultRoleID := roles[0].ID().(uuid.UUID)
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/role/%s", defaultRoleID), nil)
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, "Can't delete default roles")

	// Case 3: Success
	newRoleID := utils.Must(testApp.roleModel.CreateFromJSON(context.Background(), `{"name": "New role for delete"}`))
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/role/%v", newRoleID), nil)
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}
