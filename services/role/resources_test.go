package roleservice_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRoleServiceResources(t *testing.T) {
	testApp := createRoleTest()
	req := httptest.NewRequest("GET", "/role/resources", nil)
	req.Header.Set("Authorization", "Bearer "+testApp.adminToken)
	resp := utils.Must(testApp.server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `content.blog.create`)
}
