package schema

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilderFromDir(t *testing.T) {
	_, err := NewBuilderFromDir("../tests/invalid")
	assert.Error(t, err)

	tmpDir, err := os.MkdirTemp("../tests/", "testbuilder")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	invalidSchemaJSONFile1 := filepath.Join(tmpDir, "invalid2.json")
	utils.WriteFile(invalidSchemaJSONFile1, "{}")
	_, err = NewBuilderFromDir(tmpDir)
	assert.Error(t, err)

	invalidSchemaJSONFile2 := filepath.Join(tmpDir, "invalid1.json")
	utils.WriteFile(invalidSchemaJSONFile2, "{")
	_, err = NewBuilderFromDir(tmpDir)
	assert.Error(t, err)

	builder, err := NewBuilderFromDir("../tests/data/schemas")
	assert.Nil(t, err)
	assert.NotNil(t, builder)

	schemas := builder.Schemas()
	assert.True(t, len(schemas) > 0)

	newSchema := &Schema{
		Name: "newSchema",
	}
	schemas = append(schemas, newSchema)
	builder.AddSchema(newSchema)
	assert.Equal(t, len(schemas), len(builder.Schemas()))

	userSchema, err := builder.Schema("user")
	assert.Nil(t, err)
	assert.NotNil(t, userSchema)

	_, err = builder.Schema("invalid")
	assert.Error(t, err)
}
