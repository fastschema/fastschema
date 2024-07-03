package roleservice_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceResources(t *testing.T) {
	testApp := createTestApp()
	req := httptest.NewRequest("GET", "/api/role/resources", nil)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, response, `content.blog.create`)
}
