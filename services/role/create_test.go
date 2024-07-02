package roleservice_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceCreate(t *testing.T) {
	testApp := createTestApp()
	// Case 1: No payload
	req := httptest.NewRequest("POST", "/api/role", nil)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 2: Invalid payload data
	req = httptest.NewRequest("POST", "/api/role", bytes.NewReader([]byte(`{"name":`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, "Malformed JSON error")

	// Case 3: Invalid payload column
	req = httptest.NewRequest("POST", "/api/role", bytes.NewReader([]byte(`{"name": "New role", "invalid": "test"}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, "column role.invalid not found")

	// Case 4: Success
	req = httptest.NewRequest("POST", "/api/role", bytes.NewReader([]byte(`{"name": "New role"}`)))
	resp = utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `"id":`)
	assert.Contains(t, response, `"name":"New role"`)
}
