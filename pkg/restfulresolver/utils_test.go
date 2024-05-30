package restfulresolver_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestTransformHandlers(t *testing.T) {
	// Test case 1
	r := &fs.Resource{}
	handlers := []restfulresolver.Handler{
		func(ctx *restfulresolver.Context) error { return nil },
		func(ctx *restfulresolver.Context) error { return nil },
		func(ctx *restfulresolver.Context) error { return nil },
	}

	result1 := restfulresolver.TransformHandlers(r, handlers, nil)
	assert.Len(t, result1, 3)

	// Test case 2
	r = &fs.Resource{}
	handlers = []restfulresolver.Handler{}
	result2 := restfulresolver.TransformHandlers(r, handlers, nil)
	assert.Len(t, result2, 0)

	// Test case 3
	r = &fs.Resource{}
	handlers = []restfulresolver.Handler{
		func(ctx *restfulresolver.Context) error { return nil },
	}
	result3 := restfulresolver.TransformHandlers(r, handlers, nil)
	assert.Len(t, result3, 1)
}

func TestGetHandlerInfo(t *testing.T) {
	handler := func(ctx *restfulresolver.Context) error { return nil }
	resource := fs.NewResource("testResource", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	})

	name, handlers := restfulresolver.GetHandlerInfo(handler, nil, resource)

	assert.Equal(t, "testResource", name)
	assert.Len(t, handlers, 1)
	assert.IsType(t, (fiber.Handler)(nil), handlers[0])
}
