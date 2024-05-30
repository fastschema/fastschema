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

	entity.SetID("invalid")
	assert.Equal(t, uint64(0), entity.ID())

	entity.SetID("1")
	assert.Equal(t, uint64(1), entity.ID())

	entity.SetID(2)
	assert.Equal(t, uint64(2), entity.ID())

	entity.SetID(float64(3))
	assert.Equal(t, uint64(3), entity.ID())

	entity.SetID(uint64(4))
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

	childSlice := []*Entity{
		NewEntity().Set("id", float64(1)).Set("name", "Go"),
		NewEntity().Set("id", float64(2)).Set("name", "Python"),
		NewEntity().Set("id", float64(3)).Set("name", "Java"),
	}
	entity1.Set("skills", childSlice)

	expected := map[string]any{
		"name": "John",
		"info": map[string]any{
			"age": float64(30),
			"group": map[string]any{
				"skills": []string{"Go", "Python", "Java"},
			},
		},
		"skills": []map[string]any{
			{"id": float64(1), "name": "Go"},
			{"id": float64(2), "name": "Python"},
			{"id": float64(3), "name": "Java"},
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

	// option 1 to set slice of struct
	structElem1 := NewEntity()
	structElem1.Set("name", "name 1")
	structElem2 := NewEntity()
	structElem2.Set("name", "name 2")
	entity.Set("sliceStruct", []*Entity{structElem1, structElem2})

	// option 2 to set slice of struct
	// entity.Set("sliceStruct", []struct {
	// 	Name string `json:"name"`
	// }{
	// 	{Name: "name 1"},
	// 	{Name: "name 2"},
	// })

	type TestSliceStruct struct {
		Name string `json:"name"`
	}

	entity.Set("colors", map[string]string{
		"red":   "#ff0000",
		"green": "#00ff00",
		"blue":  "#0000ff",
	})

	// entity.Set("colors", NewEntity().
	// 	Set("red", "#f00000").
	// 	Set("green", "#00ff00").
	// 	Set("blue", "#0000ff"),
	// )

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
			{Name: "name 1"},
			{Name: "name 2"},
		},
		Colors: map[string]string{
			"red":   "#ff0000",
			"green": "#00ff00",
			"blue":  "#0000ff",
		},
	}

	result, err := BindEntity[TestStruct](entity)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestEntityToStructWithMapOfSliceOfStruct

func TestEntityToStructWithMapOfSliceOfStruct(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := NewEntity()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	entity.Set("colors", map[string]string{
		"red":   "#ff0000",
		"green": "#00ff00",
		"blue":  "#0000ff",
	})

	// entity.Set("colors", map[string][]map[string]string{
	// 	"primary": {
	// 		{"name": "Red", "code": "#ff0000"},
	// 		{"name": "Green", "code": "#00ff00"},
	// 		{"name": "Blue", "code": "#0000ff"},
	// 	},
	// 	"secondary": {
	// 		{"name": "Yellow", "code": "#ffff00"},
	// 		{"name": "Cyan", "code": "#00ffff"},
	// 		{"name": "Magenta", "code": "#ff00ff"},
	// 	},
	// })

	entity.Set("colors", NewEntity().
		Set("primary", []*Entity{
			NewEntity().Set("name", "Red").Set("code", "#ff0000"),
			NewEntity().Set("name", "Green").Set("code", "#00ff00"),
			NewEntity().Set("name", "Blue").Set("code", "#0000ff"),
		}).
		Set("secondary", []*Entity{
			NewEntity().Set("name", "Yellow").Set("code", "#ffff00"),
			NewEntity().Set("name", "Cyan").Set("code", "#00ffff"),
			NewEntity().Set("name", "Magenta").Set("code", "#ff00ff"),
		}),
	)

	type Color struct {
		Name string
		Code string
	}
	// entity.Set("colors", map[string][]Color{
	// 	"primary": {
	// 		{"Red", "#ff0000"},
	// 		{"Green", "#00ff00"},
	// 		{"Blue", "#0000ff"},
	// 	},
	// 	"secondary": {
	// 		{"Yellow", "#ffff00"},
	// 		{"Cyan", "#00ffff"},
	// 		{"Magenta", "#ff00ff"},
	// 	},
	// })

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

	result, err := BindEntity[TestStruct](entity)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestEntityToStructWithNestedStruct
func TestEntityToStructWithNestedStruct(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := NewEntity()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	colors := NewEntity()
	colors.Set("primary", []*Entity{
		NewEntity().Set("name", "Red").Set("code", "#ff0000"),
		NewEntity().Set("name", "Green").Set("code", "#00ff00"),
		NewEntity().Set("name", "Blue").Set("code", "#0000ff"),
	})
	colors.Set("secondary", []*Entity{
		NewEntity().Set("name", "Yellow").Set("code", "#ffff00"),
		NewEntity().Set("name", "Cyan").Set("code", "#00ffff"),
		NewEntity().Set("name", "Magenta").Set("code", "#ff00ff"),
	})
	entity.Set("colors", colors)

	type Color struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}

	type Group struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type TestStruct struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Skills []string `json:"skills"`
		Group  Group    `json:"group"`

		Colors struct {
			Primary   []*Color `json:"primary"`
			Secondary []*Color `json:"secondary"`
		} `json:"colors"`
	}

	expected := TestStruct{
		Name:   "John",
		Age:    30,
		Skills: []string{"Go", "Python", "Java"},
		Group: Group{
			ID:   1,
			Name: "Admin",
		},
		Colors: struct {
			Primary   []*Color `json:"primary"`
			Secondary []*Color `json:"secondary"`
		}{
			Primary: []*Color{
				{Name: "Red", Code: "#ff0000"},
				{Name: "Green", Code: "#00ff00"},
				{Name: "Blue", Code: "#0000ff"},
			},
			Secondary: []*Color{
				{Name: "Yellow", Code: "#ffff00"},
				{Name: "Cyan", Code: "#00ffff"},
				{Name: "Magenta", Code: "#ff00ff"},
			},
		},
	}

	result, err := BindEntity[TestStruct](entity)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestEntityToStructWithMissingFields(t *testing.T) {
	entity := NewEntity()
	// set missing name field
	// entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := NewEntity()

	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	// option 1 to set slice of struct
	structElem1 := NewEntity()
	structElem1.Set("name", "name 1")
	structElem2 := NewEntity()
	structElem2.Set("name", "name 2")
	entity.Set("sliceStruct", []*Entity{structElem1, structElem2})

	type TestSliceStruct struct {
		Name string `json:"name"`
	}

	entity.Set("colors", map[string]string{
		"red":   "#ff0000",
		"green": "#00ff00",
		"blue":  "#0000ff",
	})

	type TestStruct struct {
		Name        string             `json:"name,omitempty"`
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
		// Name:   "",
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
			{Name: "name 1"},
			{Name: "name 2"},
		},
		Colors: map[string]string{
			"red":   "#ff0000",
			"green": "#00ff00",
			"blue":  "#0000ff",
		},
	}

	result, err := BindEntity[TestStruct](entity)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestEntityToStructMarshalError(t *testing.T) {
	entity := NewEntity()
	entity.Set("name", make(chan int))
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	// Set an invalid value to trigger marshaling error
	entity.Set("group", make(chan int))

	type StructType struct {
		Name string `json:"name"`
	}
	_, err := BindEntity[StructType](entity)
	assert.Error(t, err)
}

func TestNamedEntity(t *testing.T) {
	entity := NamedEntity("test_entity")

	// Assert that the entity has the correct name
	assert.Equal(t, "test_entity", entity.Name())

	// Assert that the entity is not empty
	assert.True(t, entity.Empty())

	// Assert that the entity has no keys
	assert.Empty(t, entity.Keys())

	// Assert that the entity is a string representation
	assert.NotEmpty(t, entity.String())

	// Assert that the entity is the oldest key/value pair
	assert.Nil(t, entity.First())

	// Assert that the entity can set a value
	entity.Set("key", "value")
	assert.Equal(t, "value", entity.Get("key"))

	// Assert that the entity can delete a value
	entity.Delete("key")
	assert.Nil(t, entity.Get("key"))

	// Assert that the entity can get a string value
	assert.Equal(t, "default", entity.GetString("nonexistent", "default"))
}

func TestEntityName(t *testing.T) {
	entity := NamedEntity("blog")
	assert.Equal(t, "blog", entity.Name())

	entity.Name("post")
	assert.Equal(t, "post", entity.Name())
}

func TestEntityKeys(t *testing.T) {
	entity := NamedEntity("User")
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("email", "john@example.com")

	keys := entity.Keys()

	assert.Equal(t, []string{"name", "age", "email"}, keys)
}

func TestEntityKeysEmptyEntity(t *testing.T) {
	entity := NamedEntity("User")

	keys := entity.Keys()

	assert.Empty(t, keys)
}

func TestEntityKeysOrdered(t *testing.T) {
	entity := NamedEntity("User")
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("email", "john@example.com")

	keys := entity.Keys()

	assert.Equal(t, "name", keys[0])
	assert.Equal(t, "age", keys[1])
	assert.Equal(t, "email", keys[2])
}

func TestEntityKeysDuplicateKeys(t *testing.T) {
	entity := NamedEntity("User")
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("name", "Jane")

	keys := entity.Keys()

	assert.Equal(t, []string{"name", "age"}, keys)
}
