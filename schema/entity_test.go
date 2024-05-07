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

	entity.Set("group_id", "a")
	_, err = entity.GetUint64("group_id", false)
	assert.Equal(t, "invalid uint64 value group_id=a (string)", err.Error())

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

func TestEntityEmpty(t *testing.T) {
	entity := NewEntity()
	assert.True(t, entity.Empty())

	entity.Set("name", "John")
	assert.False(t, entity.Empty())
}

func TestEntityDelete(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)

	// Delete existing field
	entity.Delete("name")
	assert.Nil(t, entity.Get("name"))

	// Delete non-existing field
	entity.Delete("address")
	assert.Nil(t, entity.Get("address"))

	// Delete field with multiple values
	entity.Set("skills", []string{"Go", "Python", "Java"})
	entity.Delete("skills")
	assert.Nil(t, entity.Get("skills"))
}

func TestEntityGetString(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")

	// Test getting existing string value
	result := entity.GetString("name")
	assert.Equal(t, "John", result)

	// Test getting non-existing string value
	result = entity.GetString("age")
	assert.Equal(t, "", result)

	// Test getting existing string value with default value
	result = entity.GetString("name", "Default")
	assert.Equal(t, "John", result)

	// Test getting non-existing string value with default value
	result = entity.GetString("age", "Default")
	assert.Equal(t, "Default", result)
}

func TestNewEntityFromMap(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"age":  30,
		"skills": []string{
			"Go",
			"Python",
			"Java",
		},
		"group": map[string]any{
			"id":   1,
			"name": "Admin",
		},
	}

	entity := NewEntityFromMap(data)

	assert.Equal(t, "John", entity.Get("name"))
	assert.Equal(t, 30, entity.Get("age"))
	assert.Equal(t, []string{"Go", "Python", "Java"}, entity.Get("skills"))

	group := entity.Get("group")
	assert.NotNil(t, group)

	groupEntity, ok := group.(*Entity)
	assert.True(t, ok)

	assert.Equal(t, 1, groupEntity.Get("id"))
	assert.Equal(t, "Admin", groupEntity.Get("name"))
}

func TestToJsonError(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")

	_, err := entity.ToJSON()
	assert.NoError(t, err)

	// Test error when marshaling entity to JSON
	entity.Set("skills", make(chan int))
	_, err = entity.ToJSON()
	assert.Error(t, err)
}
func TestEntityToMap(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})
	group := NewEntity()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	expected := map[string]any{
		"name": "John",
		"age":  30,
		"skills": []string{
			"Go",
			"Python",
			"Java",
		},
		"group": map[string]any{
			"id":   1,
			"name": "Admin",
		},
	}

	result := entity.ToMap()
	assert.Equal(t, expected, result)
}

func TestEntityToMapEmptyEntity(t *testing.T) {
	entity := NewEntity()

	expected := map[string]any{}

	result := entity.ToMap()
	assert.Equal(t, expected, result)
}

func TestEntityToMapNestedEntities(t *testing.T) {
	entity1 := NewEntity()
	entity1.Set("name", "John")

	entity2 := NewEntity()
	entity2.Set("age", float64(30))

	entity3 := NewEntity()
	entity3.Set("skills", []string{"Go", "Python", "Java"})

	entity2.Set("group", entity3)

	entity1.Set("info", entity2)

	expected := map[string]any{
		"name": "John",
		"info": map[string]any{
			"age": float64(30),
			"group": map[string]any{
				"skills": []string{"Go", "Python", "Java"},
			},
		},
	}

	result := entity1.ToMap()
	assert.Equal(t, expected, result)
}

func TestEntityUnmarshalJSON(t *testing.T) {
	entity := NewEntity()

	t.Run("ObjectEach", func(t *testing.T) {
		data := []byte(`{
			"name": "John",
			"age": 30,
			"skills": [
				{
					"id": 1,
					"name": "Go"
				},
				{
					"id": 2,
					"name": "Python"
				}
			]
		}`)

		err := entity.UnmarshalJSON(data)
		assert.NoError(t, err)

		assert.Equal(t, "John", entity.Get("name"))
		assert.Equal(t, float64(30), entity.Get("age"))

		skills := entity.Get("skills")
		assert.NotNil(t, skills)

		skillsArr, ok := skills.([]*Entity)
		assert.True(t, ok)
		assert.Equal(t, 2, len(skillsArr))

		assert.Equal(t, float64(1), skillsArr[0].Get("id"))
		assert.Equal(t, "Go", skillsArr[0].Get("name"))

		assert.Equal(t, float64(2), skillsArr[1].Get("id"))
		assert.Equal(t, "Python", skillsArr[1].Get("name"))
	})

	t.Run("ArrayEach", func(t *testing.T) {
		data := []byte(`{
			"numbers": [1, 2, 3, 4, 5]
		}`)

		err := entity.UnmarshalJSON(data)
		assert.NoError(t, err)

		numbers := entity.Get("numbers")
		assert.NotNil(t, numbers)

		numbersArr, ok := numbers.([]any)
		assert.True(t, ok)
		assert.Equal(t, []any{float64(1), float64(2), float64(3), float64(4), float64(5)}, numbersArr)
	})

	t.Run("MixedEntitiesAndNonEntities", func(t *testing.T) {
		data := []byte(`{"mixed":[{"id":1,"name":"Go"},"Python","Java"]}`)

		err := entity.UnmarshalJSON(data)
		assert.Error(t, err)
		assert.Equal(t, "cannot mix entities and non-entities in a slice: mixed=[{\"id\":1,\"name\":\"Go\"},\"Python\",\"Java\"]", err.Error())
	})
}

func TestEntityUnmarshalJSONErrorArrayEach(t *testing.T) {
	entity := NewEntity()

	data := []byte(`{
		"mixed": [1, 2, 3,]
	}`)

	err := entity.UnmarshalJSON(data)
	assert.Error(t, err)
	assert.Equal(t, "Unknown value type", err.Error())
}

func TestEntityToStruct(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := NewEntity()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	type TestSliceStruct struct {
		Name string `json:"name"`
	}

	sliceStruct1 := NewEntity()
	sliceStruct1.Set("name", "name 1")
	sliceStruct2 := NewEntity()
	sliceStruct2.Set("name", "name 2")

	entity.Set("sliceStruct", []*Entity{
		sliceStruct1,
		sliceStruct2,
	})

	entity.Set("colors", map[string]string{
		"red":   "#ff0000",
		"green": "#00ff00",
		"blue":  "#0000ff",
	})

	type TestStruct struct {
		Name        string             `json:"name"`
		Age         int                `json:"age"`
		Skills      []string           `json:"skills"`
		SliceStruct []*TestSliceStruct `json:"sliceStruct"`
		Group       struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"group"`

		Colors map[string]string `json:"colors"`
	}

	expected := TestStruct{
		Name:   "John",
		Age:    30,
		Skills: []string{"Go", "Python", "Java"},
		Group: struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{
			ID:   1,
			Name: "Admin",
		},
		SliceStruct: []*TestSliceStruct{
			&TestSliceStruct{Name: "name 1"},
			&TestSliceStruct{Name: "name 2"},
		},
		Colors: map[string]string{
			"red":   "#ff0000",
			"green": "#00ff00",
			"blue":  "#0000ff",
		},
	}

	result := entity.EntityToStruct(&TestStruct{})

	assert.Equal(t, expected, result)
}

func TestEntityToStructWithMapOfSliceOfStruct(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := NewEntity()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	type Color struct {
		Name string
		Code string
	}

	entity.Set("colors", map[string][]*Color{
		"primary": {
			{"Red", "#ff0000"},
			{"Green", "#00ff00"},
			{"Blue", "#0000ff"},
		},
		"secondary": {
			{"Yellow", "#ffff00"},
			{"Cyan", "#00ffff"},
			{"Magenta", "#ff00ff"},
		},
	})

	type TestStruct struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Skills []string `json:"skills"`
		Group  struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"group"`

		Colors map[string][]*Color `json:"colors"`
	}

	expected := TestStruct{
		Name:   "John",
		Age:    30,
		Skills: []string{"Go", "Python", "Java"},
		Group: struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{
			ID:   1,
			Name: "Admin",
		},
		Colors: map[string][]*Color{
			"primary": {
				{"Red", "#ff0000"},
				{"Green", "#00ff00"},
				{"Blue", "#0000ff"},
			},
			"secondary": {
				{"Yellow", "#ffff00"},
				{"Cyan", "#00ffff"},
				{"Magenta", "#ff00ff"},
			},
		},
	}

	result := entity.EntityToStruct(&TestStruct{})

	assert.Equal(t, expected, result)
}

func TestEntityToStructWithNestedStruct(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})
	group := NewEntity()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	type Group struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type TestStruct struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Skills []string `json:"skills"`
		Group  Group    `json:"group"`
	}

	expected := TestStruct{
		Name:   "John",
		Age:    30,
		Skills: []string{"Go", "Python", "Java"},
		Group: Group{
			ID:   1,
			Name: "Admin",
		},
	}

	result := entity.EntityToStruct(&TestStruct{})

	assert.Equal(t, expected, result)
}

func TestEntityToStructWithMissingFields(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")

	type TestStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Skills []string
		Group  struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"group"`
	}

	expected := TestStruct{
		Name:   "John",
		Age:    0,
		Skills: nil,
		Group: struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{
			ID:   0,
			Name: "",
		},
	}

	result := entity.EntityToStruct(&TestStruct{})

	assert.Equal(t, expected, result)
}

func TestEntityToStructWithInvalidNestedStruct(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "reflect.Set: value of type string is not assignable to type int", r)
		}
	}()

	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})
	group := NewEntity()
	group.Set("id", "invalid")
	group.Set("name", "Admin")
	entity.Set("group", group)

	type Group struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type TestStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Skills []string
		Group  Group `json:"group"`
	}

	_ = entity.EntityToStruct(&TestStruct{})
}
