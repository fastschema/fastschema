package entity

import (
	"testing"

	"github.com/buger/jsonparser"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestEntity(t *testing.T) {
	entity := New()

	entity.Set("name", "John")
	assert.Equal(t, "John", entity.Get("name"))

	assert.Equal(t, "defaultLocation", entity.Get("location", "defaultLocation"))
	assert.Equal(t, nil, entity.Get("location"))

	assert.Equal(t, "John", entity.First().Value)
	assert.Nil(t, entity.ID())

	entity.Set("id", "invalid")
	assert.Equal(t, "invalid", entity.ID())

	assert.Equal(t, uint64(0), utils.Must(entity.GetUint64("group_id", true)))
	_, err := entity.GetUint64("group_id", false)
	assert.Equal(t, "cannot get uint64 value from entity: group_id", err.Error())

	entity.Set("group_id", uint64(1))
	assert.Equal(t, uint64(1), utils.Must(entity.GetUint64("group_id", false)))

	entity.Set("group_id", "a")
	_, err = entity.GetUint64("group_id", false)
	assert.Contains(t, err.Error(), "cannot get uint64 value from entity")

	entity.Set("group_id", 1)
	assert.Equal(t, 1, entity.Get("group_id"))

	assert.Error(t, entity.SetID(nil))

	assert.NoError(t, entity.SetID("1"))
	assert.Equal(t, "1", entity.ID())

	assert.NoError(t, entity.SetID(2))
	assert.Equal(t, 2, entity.ID())

	assert.NoError(t, entity.SetID(float64(3)))
	assert.Equal(t, float64(3), entity.ID())

	assert.NoError(t, entity.SetID(uint64(4)))
	assert.Equal(t, uint64(4), entity.ID())

	entityBytes, err := entity.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"John","id":4,"group_id":1}`, string(entityBytes))

	entity2 := New()
	assert.NoError(t, entity2.UnmarshalJSON(entityBytes))
	assert.Equal(t, entity.Get("name"), entity2.Get("name"))

	entity3 := New()
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

	assert.Equal(t, float64(4), entity3.ID())
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

	assert.Equal(t, jsonString, entity3.String())

	entity4 := New(4)
	assert.Equal(t, 4, entity4.ID())

	entity5, err := NewEntityFromJSON(`{"id":5,"name":"John","group_id":1,"group":{"id":1,"name":"Admin"},"tags":["developer","admin"],"skills":[{"id":1,"name":"Go"},{"id":2,"name":"PHP"}]}`)
	assert.NoError(t, err)
	assert.Equal(t, float64(5), entity5.ID())

	_, err = NewEntityFromJSON(`{"id":5`)
	assert.Error(t, err)

	data := entity5.Data()
	assert.Equal(t, 6, data.Len())

	// Stringify entity with invalid value
	entity5.Set("skills", make(chan int))
	str := entity5.String()
	assert.Contains(t, str, "cannot convert entity to string")
}

func TestEntitySetIDCustomField(t *testing.T) {
	e := New()
	e.SetIDField("code")
	assert.NoError(t, e.SetID("proj-9"))
	assert.Equal(t, "proj-9", e.ID())
	assert.Equal(t, "proj-9", e.Get("code"))
	assert.Equal(t, "proj-9", e.Get(FieldID))
}

func TestEntityEmpty(t *testing.T) {
	entity := New()
	assert.True(t, entity.Empty())

	entity.Set("name", "John")
	assert.False(t, entity.Empty())
}

func TestEntityDelete(t *testing.T) {
	entity := New()
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
	entity := New()
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
	entity := New()
	entity.Set("name", "John")

	_, err := entity.ToJSON()
	assert.NoError(t, err)

	// Test error when marshaling entity to JSON
	entity.Set("skills", make(chan int))
	_, err = entity.ToJSON()
	assert.Error(t, err)
}
func TestEntityToMap(t *testing.T) {
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})
	group := New()
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
	entity := New()

	expected := map[string]any{}

	result := entity.ToMap()
	assert.Equal(t, expected, result)
}

func TestEntityToMapNestedEntities(t *testing.T) {
	entity1 := New()
	entity1.Set("name", "John")

	entity2 := New()
	entity2.Set("age", float64(30))

	entity3 := New()
	entity3.Set("skills", []string{"Go", "Python", "Java"})

	entity2.Set("group", entity3)

	entity1.Set("info", entity2)

	childSlice := []*Entity{
		New().Set("id", float64(1)).Set("name", "Go"),
		New().Set("id", float64(2)).Set("name", "Python"),
		New().Set("id", float64(3)).Set("name", "Java"),
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
	entity := New()

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
	entity := New()

	data := []byte(`{
		"mixed": [1, 2, 3,]
	}`)

	err := entity.UnmarshalJSON(data)
	assert.Error(t, err)
	assert.Equal(t, "Unknown value type", err.Error())
}

func TestEntityToStruct(t *testing.T) {
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := New()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	// option 1 to set slice of struct
	structElem1 := New()
	structElem1.Set("name", "name 1")
	structElem2 := New()
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
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := New()
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

	entity.Set("colors", New().
		Set("primary", []*Entity{
			New().Set("name", "Red").Set("code", "#ff0000"),
			New().Set("name", "Green").Set("code", "#00ff00"),
			New().Set("name", "Blue").Set("code", "#0000ff"),
		}).
		Set("secondary", []*Entity{
			New().Set("name", "Yellow").Set("code", "#ffff00"),
			New().Set("name", "Cyan").Set("code", "#00ffff"),
			New().Set("name", "Magenta").Set("code", "#ff00ff"),
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
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := New()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	colors := New()
	colors.Set("primary", []*Entity{
		New().Set("name", "Red").Set("code", "#ff0000"),
		New().Set("name", "Green").Set("code", "#00ff00"),
		New().Set("name", "Blue").Set("code", "#0000ff"),
	})
	colors.Set("secondary", []*Entity{
		New().Set("name", "Yellow").Set("code", "#ffff00"),
		New().Set("name", "Cyan").Set("code", "#00ffff"),
		New().Set("name", "Magenta").Set("code", "#ff00ff"),
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
	entity := New()
	// set missing name field
	// entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := New()

	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	// option 1 to set slice of struct
	structElem1 := New()
	structElem1.Set("name", "name 1")
	structElem2 := New()
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
	entity := New()
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

func TestEntityKeys(t *testing.T) {
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("email", "john@example.com")

	keys := entity.Keys()

	assert.Equal(t, []string{"name", "age", "email"}, keys)
}

func TestEntityKeysEmptyEntity(t *testing.T) {
	entity := New()

	keys := entity.Keys()

	assert.Empty(t, keys)
}

func TestEntityKeysOrdered(t *testing.T) {
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("email", "john@example.com")

	keys := entity.Keys()

	assert.Equal(t, "name", keys[0])
	assert.Equal(t, "age", keys[1])
	assert.Equal(t, "email", keys[2])
}

func TestEntityKeysDuplicateKeys(t *testing.T) {
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("name", "Jane")

	keys := entity.Keys()

	assert.Equal(t, []string{"name", "age"}, keys)
}
func TestBindEntity(t *testing.T) {
	entity := New()
	entity.Set("name", "John")
	entity.Set("age", 30)
	entity.Set("skills", []string{"Go", "Python", "Java"})

	group := New()
	group.Set("id", 1)
	group.Set("name", "Admin")
	entity.Set("group", group)

	type TestStruct struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Skills []string `json:"skills"`
		Group  struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"group"`
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
	}

	result, err := BindEntity[TestStruct](entity)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	errEntity := New()
	errEntity.Set("name", "John")
	errEntity.Set("age", make(chan int))

	_, err = BindEntity[TestStruct](errEntity)
	assert.Error(t, err)
}

// TestNewWithIDField tests the NewWithIDField constructor function.
func TestNewWithIDField(t *testing.T) {
	// Test with custom ID field and ID value
	e := NewWithIDField("code", "proj-123")
	assert.Equal(t, "code", e.GetIDField())
	assert.Equal(t, "proj-123", e.ID())
	assert.Equal(t, "proj-123", e.Get("code"))
	assert.Equal(t, "proj-123", e.Get(FieldID))

	// Test with custom ID field but no ID value
	e2 := NewWithIDField("sku")
	assert.Equal(t, "sku", e2.GetIDField())
	assert.Nil(t, e2.ID())

	// Test with empty field name (should default to "id")
	e3 := NewWithIDField("", 42)
	assert.Equal(t, FieldID, e3.GetIDField())
	assert.Equal(t, 42, e3.ID())
}

// TestSetIDFieldEdgeCases tests edge cases for SetIDField.
func TestSetIDFieldEdgeCases(t *testing.T) {
	// Test nil entity receiver
	var nilEntity *Entity
	result := nilEntity.SetIDField("custom")
	assert.Nil(t, result)

	// Test empty field name defaults to FieldID
	e := New()
	e.SetIDField("custom")
	assert.Equal(t, "custom", e.GetIDField())
	e.SetIDField("")
	assert.Equal(t, FieldID, e.GetIDField())
}

// TestGetIDFieldEdgeCases tests edge cases for GetIDField.
func TestGetIDFieldEdgeCases(t *testing.T) {
	// Test nil entity receiver
	var nilEntity *Entity
	assert.Equal(t, FieldID, nilEntity.GetIDField())

	// Test entity with empty idField (should default to FieldID)
	e := New()
	// Manually clear the idField to test fallback
	e.idField = ""
	assert.Equal(t, FieldID, e.GetIDField())
}

// TestIDFallbackPath tests the ID() method's fallback behavior.
func TestIDFallbackPath(t *testing.T) {
	// Test when custom idField is set but value not found, falls back to FieldID
	e := New()
	e.idField = "custom_id"
	e.Set(FieldID, 999) // Set only the default field
	// ID() should first look for "custom_id", not find it, then fall back to "id"
	assert.Equal(t, 999, e.ID())

	// Test when neither custom field nor FieldID has value
	e2 := New()
	e2.idField = "custom_id"
	assert.Nil(t, e2.ID())

	// Test when idField equals FieldID (no fallback needed)
	e3 := New()
	e3.Set(FieldID, 123)
	assert.Equal(t, 123, e3.ID())
}

// TestUnmarshalJSONEmptyArray tests unmarshaling an empty JSON array.
func TestUnmarshalJSONEmptyArray(t *testing.T) {
	entity := New()
	err := entity.UnmarshalJSON([]byte(`{"items": []}`))
	assert.NoError(t, err)
	assert.Nil(t, entity.Get("items"))
}

// TestGetStringNonStringValue tests GetString with a non-string value.
func TestGetStringNonStringValue(t *testing.T) {
	entity := New()
	entity.Set("count", 42)

	// Should return empty string when value is not a string and no default
	result := entity.GetString("count")
	assert.Equal(t, "", result)

	// Should return default when value is not a string
	result = entity.GetString("count", "default")
	assert.Equal(t, "default", result)
}
