package restresolver_test

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/stretchr/testify/assert"
)

func TestTransformHandlers(t *testing.T) {
	// Test case 1
	r := &app.Resource{}
	handlers := []restresolver.Handler{
		func(ctx *restresolver.Context) error { return nil },
		func(ctx *restresolver.Context) error { return nil },
		func(ctx *restresolver.Context) error { return nil },
	}

	result1 := restresolver.TransformHandlers(r, handlers, nil)
	assert.Len(t, result1, 3)

	// Test case 2
	r = &app.Resource{}
	handlers = []restresolver.Handler{}
	result2 := restresolver.TransformHandlers(r, handlers, nil)
	assert.Len(t, result2, 0)

	// Test case 3
	r = &app.Resource{}
	handlers = []restresolver.Handler{
		func(ctx *restresolver.Context) error { return nil },
	}
	result3 := restresolver.TransformHandlers(r, handlers, nil)
	assert.Len(t, result3, 1)
}
