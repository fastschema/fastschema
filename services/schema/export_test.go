package schemaservice_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSchemaServiceExport(t *testing.T) {
	_, _, server := createSchemaService(t, nil)

	// Case 1: schema already exists
	req := httptest.NewRequest("GET", "/schema/export/blog", nil)

	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema blog not found`)

	// Case 2: export success
	req = httptest.NewRequest("GET", "/schema/export/category", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}
