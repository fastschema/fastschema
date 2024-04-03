package schema

import (
	"testing"

	"github.com/buger/jsonparser"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestEntity(t *testing.T) {
	entity := NewEntity()

	entity.Set("name", "John")
	assert.Equal(t, "John", entity.Get("name"))

	assert.Equal(t, "defaultLocation", entity.Get("location", "defaultLocation"))
	assert.Equal(t, nil, entity.Get("location"))

	assert.Equal(t, "John", entity.First().Value)
	assert.Equal(t, uint64(0), entity.ID())

	entity.Set("id", "invalid")
	assert.Equal(t, uint64(0), entity.ID())

	assert.Equal(t, uint64(0), utils.Must(entity.GetUint64("group_id", true)))
	_, err := entity.GetUint64("group_id", false)
	assert.Equal(t, "cannot get uint64 value from entity: group_id", err.Error())

	entity.Set("group_id", uint64(1))
	assert.Equal(t, uint64(1), utils.Must(entity.GetUint64("group_id", false)))

	entity.Set("group_id", "1")
	_, err = entity.GetUint64("group_id", false)
	assert.Equal(t, "invalid uint64 value group_id=1 (string)", err.Error())

	entity.Set("group_id", 1)
	assert.Equal(t, 1, entity.Get("group_id"))

	assert.Error(t, entity.SetID("invalid"))
	assert.NoError(t, entity.SetID("1"))
	assert.Equal(t, uint64(1), entity.ID())

	assert.NoError(t, entity.SetID(2))
	assert.Equal(t, uint64(2), entity.ID())

	assert.NoError(t, entity.SetID(float64(3)))
	assert.Equal(t, uint64(3), entity.ID())

	assert.NoError(t, entity.SetID(uint64(4)))
	assert.Equal(t, uint64(4), entity.ID())

	entityBytes, err := entity.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"John","id":4,"group_id":1}`, string(entityBytes))

	entity2 := NewEntity()
	assert.NoError(t, entity2.UnmarshalJSON(entityBytes))
	assert.Equal(t, entity.Get("name"), entity2.Get("name"))

	entity3 := NewEntity()
	assert.Error(t, entity3.UnmarshalJSON([]byte(`{`)))
	assert.Error(t, entity3.UnmarshalJSON([]byte(`{
		"id":4,"name":"John",
		"skills": [
			{
				"id:1,
			}
		]
	}`)))
	assert.Error(t, entity3.UnmarshalJSON([]byte(`{
		"id":4,"name":"John",
		"skills": [
			"go",
			{
				"id": 1,
			}
		]
	}`)))

	assert.NoError(t, entity3.UnmarshalJSON([]byte(`{
		"id":4,
		"name":"John",
		"group_id":1,
		"group": {
			"id":1,
			"name":"Admin"
		},
		"tags": [
			"developer",
			"admin"
		],
		"skills": [
			{
				"id":1,
				"name":"Go"
			},
			{
				"id":2,
				"name":"PHP"
			}
		]
	}`)))

	assert.Equal(t, uint64(4), entity3.ID())
	assert.Equal(t, "John", entity3.Get("name"))

	value, err := UnmarshalJSONValue(nil, []byte("name"), jsonparser.String, 0)
	assert.NoError(t, err)
	assert.Equal(t, "name", value)

	value, err = UnmarshalJSONValue(nil, []byte("1"), jsonparser.Number, 0)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), value)

	value, err = UnmarshalJSONValue(nil, []byte("false"), jsonparser.Boolean, 0)
	assert.NoError(t, err)
	assert.Equal(t, false, value)

	_, err = UnmarshalJSONValue(nil, []byte("name"), jsonparser.Boolean, 0)
	assert.Error(t, err)

	jsonString, err := entity3.ToJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"id":4,"name":"John","group_id":1,"group":{"id":1,"name":"Admin"},"tags":["developer","admin"],"skills":[{"id":1,"name":"Go"},{"id":2,"name":"PHP"}]}`, jsonString)

	entity4 := NewEntity(4)
	assert.Equal(t, uint64(4), entity4.ID())

	entity5, err := NewEntityFromJSON(`{"id":5,"name":"John","group_id":1,"group":{"id":1,"name":"Admin"},"tags":["developer","admin"],"skills":[{"id":1,"name":"Go"},{"id":2,"name":"PHP"}]}`)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), entity5.ID())

	_, err = NewEntityFromJSON(`{"id":5`)
	assert.Error(t, err)
}
