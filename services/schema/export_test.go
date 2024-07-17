package schemaservice_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSchemaServiceExport(t *testing.T) {
	_, _, server := createSchemaService(t, nil)

	// Case 1: schema is not exists
	req := httptest.NewRequest("POST", "/schema/export", bytes.NewReader([]byte(`{"schemas":["blog"]}`)))

	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 404, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schema blog not found`)

	// Case 2: schemas is empty
	req = httptest.NewRequest("POST", "/schema/export", bytes.NewReader([]byte(`{"schemas":[]}`)))

	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.Contains(t, response, `schemas is required`)

	// Case 3: export success
	req = httptest.NewRequest("POST", "/schema/export", bytes.NewReader([]byte(`{"schemas":["category"]}`)))
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}
