package restresolver_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func closeResponse(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	assert.NoError(t, err)
}
