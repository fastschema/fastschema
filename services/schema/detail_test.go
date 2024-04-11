package schemaservice_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSchemaServiceDetail(t *testing.T) {
	_, _, server := createSchemaService(t, nil)

	// Case 1: schema not found
	req := httptest.NewRequest("GET", "/schema/product", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema product not found`)

	// Case 2: scuccess
	req = httptest.NewRequest("GET", "/schema/category", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `category`)
	assert.Contains(t, response, `categories`)
	assert.Contains(t, response, `name`)
	assert.Contains(t, response, `Name`)
	assert.Contains(t, response, `sortable`)
}
