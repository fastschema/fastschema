package schema_test

import (
	"testing"
	"time"

	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func createField(fieldType schema.FieldType, name, label string) *schema.Field {
	return &schema.Field{
		IsSystemField: true,
		Type:          fieldType,
		Name:          name,
		Label:         label,
	}
}

func createFieldWithDefault(fieldType schema.FieldType, name, label string, defaultValue any) *schema.Field {
	f := createField(fieldType, name, label)
	f.Default = defaultValue
	return f
}

func TestCreateSchemaErrorNonStruct(t *testing.T) {
	types := []struct {
		t   any
		err string
	}{
		{t: nil, err: "<nil>"},
		{t: 1, err: "int"},
		{t: "string", err: "string"},
	}

	for _, tt := range types {
		s, err := schema.CreateSchema(tt.t)
		assert.Nil(t, s)
		assert.EqualError(t, err, "can not create schema from invalid type "+tt.err)
	}
}

func TestCreateSchemaFieldTagCommon(t *testing.T) {
	type Category struct {
		Name string `json:"title" fs:"label=Category Name;multiple;unique;optional;sortable;filterable;size=255"`
		Slug string `json:"slug" fs:"label_field"`
	}

	ss, err := schema.CreateSchema(Category{})
	assert.NoError(t, err)
	assert.Equal(t, "category", ss.Name)
	assert.Equal(t, "categories", ss.Namespace)

	expectedFields := []*schema.Field{
		{
			IsSystemField: true,
			Type:          schema.TypeString,
			Name:          "title",
			Label:         "Category Name",
			IsMultiple:    true,
			Unique:        true,
			Optional:      true,
			Sortable:      true,
			Filterable:    true,
			Size:          255,
		},
		{
			IsSystemField: true,
			Type:          schema.TypeString,
			Name:          "slug",
			Label:         "Slug",
		},
	}

	assert.Equal(t, expectedFields, ss.Fields)
	assert.Equal(t, "slug", ss.LabelFieldName)
}

func TestCreateSchemaFieldTagEnums(t *testing.T) {
	// Case 1: Invalid enum values
	type Category1 struct {
		Name     string `json:"name"`
		Statuses string `json:"statuses" fs.enums:"invalid"`
	}

	ss, err := schema.CreateSchema(Category1{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid enums format")

	// Case 2: Success
	type Category2 struct {
		Name     string `json:"name"`
		Statuses string `json:"statuses" fs.enums:"[{'value':'active','label':'Active'},{'value':'inactive','label':'Inactive'}]"`
	}

	ss, err = schema.CreateSchema(Category2{})
	assert.NoError(t, err)
	expectedFields := []*schema.Field{
		{
			IsSystemField: true,
			Type:          schema.TypeString,
			Name:          "name",
			Label:         "Name",
		},
		{
			IsSystemField: true,
			Type:          schema.TypeEnum,
			Name:          "statuses",
			Label:         "Statuses",
			Enums: []*schema.FieldEnum{
				{Value: "active", Label: "Active"},
				{Value: "inactive", Label: "Inactive"},
			},
		},
	}

	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaFieldTagRelation(t *testing.T) {
	// Case 1: Invalid enum values
	type Category1 struct {
		Name  string `json:"name"`
		Posts string `json:"statuses" fs.relation:"invalid"`
	}

	ss, err := schema.CreateSchema(Category1{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid relation format")

	// Case 2: Success
	type Category2 struct {
		Name  string `json:"name"`
		Posts string `json:"posts" fs.relation:"{'type':'o2m','schema':'post','field':'category','owner':true,'optional':true}"`
	}

	ss, err = schema.CreateSchema(Category2{})
	assert.NoError(t, err)
	expectedFields := []*schema.Field{
		{
			IsSystemField: true,
			Type:          schema.TypeString,
			Name:          "name",
			Label:         "Name",
		},
		{
			IsSystemField: true,
			Type:          schema.TypeRelation,
			Name:          "posts",
			Label:         "Posts",
			Relation: &schema.Relation{
				Type:             schema.O2M,
				TargetSchemaName: "post",
				TargetFieldName:  "category",
				Owner:            true,
				Optional:         true,
			},
		},
	}

	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaFieldTagRenderer(t *testing.T) {
	// Case 1: Invalid renderer
	type Category1 struct {
		Name string `json:"name" fs.renderer:"indalid"`
	}

	ss, err := schema.CreateSchema(Category1{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid renderer format")

	// Case 2: Success
	type Category2 struct {
		Name string `json:"name" fs.renderer:"{'class': 'TitleRenderer', 'settings': {'size': 255}}"`
	}

	ss, err = schema.CreateSchema(Category2{})
	assert.NoError(t, err)
	expectedFields := []*schema.Field{
		{
			IsSystemField: true,
			Type:          schema.TypeString,
			Name:          "name",
			Label:         "Name",
			Renderer: &schema.FieldRenderer{
				Class: "TitleRenderer",
				Settings: map[string]any{
					"size": float64(255),
				},
			},
		},
	}

	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaFieldTagDB(t *testing.T) {
	// Case 1: Invalid db
	type Category1 struct {
		Name string `json:"name" fs.db:"indalid"`
	}

	ss, err := schema.CreateSchema(Category1{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid db format")

	// Case 2: Success
	type Category2 struct {
		Name string `json:"name" fs.db:"{'attr':'unique','collation':'utf8mb4_unicode_ci','increment':true,'key':'PRI'}"`
	}

	ss, err = schema.CreateSchema(Category2{})
	assert.NoError(t, err)
	expectedFields := []*schema.Field{
		{
			IsSystemField: true,
			Type:          schema.TypeString,
			Name:          "name",
			Label:         "Name",
			DB: &schema.FieldDB{
				Attr:      "unique",
				Collation: "utf8mb4_unicode_ci",
				Increment: true,
				Key:       "PRI",
			},
		},
	}

	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaFieldTagDefault(t *testing.T) {
	// Case 1: Default valid value
	type Category struct {
		String               string     `json:"string" fs:"default=hello"`
		Bool                 bool       `json:"bool" fs:"default=true"`
		Int                  int        `json:"int" fs:"default=10"`
		Int8                 int8       `json:"int8" fs:"default=10"`
		Int16                int16      `json:"int16" fs:"default=10"`
		Int32                int32      `json:"int32" fs:"default=10"`
		Int64                int64      `json:"int64" fs:"default=10"`
		Uint                 uint       `json:"uint" fs:"default=10"`
		Uint8                uint8      `json:"uint8" fs:"default=10"`
		Uint16               uint16     `json:"uint16" fs:"default=10"`
		Uint32               uint32     `json:"uint32" fs:"default=10"`
		Uint64               uint64     `json:"uint64" fs:"default=10"`
		Float32              float32    `json:"float32" fs:"default=10.5"`
		Float64              float64    `json:"float64" fs:"default=10.5"`
		Time1                time.Time  `json:"time1" fs:"default=2024-05-19T16:45:01Z"`
		Time2                *time.Time `json:"time2" fs:"default=2024-05-19T16:45:01-07:00"`
		IgnoreFieldByJSONTag string     `json:"-"`
		IgnoreFieldByFSTag   string     `fs:"-"`
	}

	ss, err := schema.CreateSchema(Category{})
	assert.NoError(t, err)
	assert.Equal(t, "category", ss.Name)
	assert.Equal(t, "categories", ss.Namespace)

	expectedFields := []*schema.Field{
		createFieldWithDefault(schema.TypeString, "string", "String", "hello"),
		createFieldWithDefault(schema.TypeBool, "bool", "Bool", true),
		createFieldWithDefault(schema.TypeInt, "int", "Int", 10),
		createFieldWithDefault(schema.TypeInt8, "int8", "Int8", int8(10)),
		createFieldWithDefault(schema.TypeInt16, "int16", "Int16", int16(10)),
		createFieldWithDefault(schema.TypeInt32, "int32", "Int32", int32(10)),
		createFieldWithDefault(schema.TypeInt64, "int64", "Int64", int64(10)),
		createFieldWithDefault(schema.TypeUint, "uint", "Uint", uint(10)),
		createFieldWithDefault(schema.TypeUint8, "uint8", "Uint8", uint8(10)),
		createFieldWithDefault(schema.TypeUint16, "uint16", "Uint16", uint16(10)),
		createFieldWithDefault(schema.TypeUint32, "uint32", "Uint32", uint32(10)),
		createFieldWithDefault(schema.TypeUint64, "uint64", "Uint64", uint64(10)),
		createFieldWithDefault(schema.TypeFloat32, "float32", "Float32", float32(10.5)),
		createFieldWithDefault(schema.TypeFloat64, "float64", "Float64", 10.5),
		createFieldWithDefault(schema.TypeTime, "time1", "Time1", time.Date(2024, 5, 19, 16, 45, 1, 0, time.UTC)),
		createFieldWithDefault(schema.TypeTime, "time2", "Time2", time.Date(2024, 5, 19, 16, 45, 1, 0, time.FixedZone("", -7*60*60))),
	}
	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaFieldTagError(t *testing.T) {
	// Case 1: Invalid field type
	type Category1 struct {
		String string `fs:"type=invalid"`
	}
	ss, err := schema.CreateSchema(Category1{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid field type")

	// Case 1: Invalid field size
	type Category2 struct {
		String string `fs:"size=invalid"`
	}
	ss, err = schema.CreateSchema(Category2{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid field size")

	// Case 3: Default invalid value
	type Category3 struct {
		String string `json:"string" fs:"default=hello"`
		Bool   bool   `json:"bool" fs:"default=hello"`
	}
	ss, err = schema.CreateSchema(Category3{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "category_3 invalid field value")

	// Case 4: Invalid enum value format
	type Category5 struct {
		Enum string `json:"enum" fs.enums:"enums=k1"`
	}
	ss, err = schema.CreateSchema(Category5{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid enums format")

	// Case 5: Invalid relation value format
	type Category6 struct {
		Name     string `json:"name"`
		Relation string `json:"relation" fs.relation:"invalid"`
	}
	ss, err = schema.CreateSchema(Category6{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid relation format")

	// Case 7: Invalid relation type
	type Category7 struct {
		Name  string       `json:"name"`
		Posts []*Category6 `json:"relation" fs.relation:"{'schema': 'post', 'field': 'category', 'type': 'invalid'}"`
	}
	ss, err = schema.CreateSchema(Category7{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "relation type is required")
}

func TestCreateSchemaWithSimpleField(t *testing.T) {
	type Category struct {
		Bool    bool       `json:"bool"`
		String  string     `json:"string"`
		Int     int        `json:"int"`
		Int8    int8       `json:"int8"`
		Int16   int16      `json:"int16"`
		Int32   int32      `json:"int32"`
		Int64   int64      `json:"int64"`
		Uint    uint       `json:"uint"`
		Uint8   uint8      `json:"uint8"`
		Uint16  uint16     `json:"uint16"`
		Uint32  uint32     `json:"uint32"`
		Uint64  uint64     `json:"uint64"`
		Float32 float32    `json:"float32"`
		Float64 float64    `json:"float64"`
		Time1   time.Time  `json:"time1"`
		Time2   *time.Time `json:"time2"`
		Text    string     `json:"text" fs:"type=text"`
	}

	ss, err := schema.CreateSchema(Category{})
	assert.NoError(t, err)
	assert.Equal(t, "category", ss.Name)
	assert.Equal(t, "categories", ss.Namespace)

	expectedFields := []*schema.Field{
		createField(schema.TypeBool, "bool", "Bool"),
		createField(schema.TypeString, "string", "String"),
		createField(schema.TypeInt, "int", "Int"),
		createField(schema.TypeInt8, "int8", "Int8"),
		createField(schema.TypeInt16, "int16", "Int16"),
		createField(schema.TypeInt32, "int32", "Int32"),
		createField(schema.TypeInt64, "int64", "Int64"),
		createField(schema.TypeUint, "uint", "Uint"),
		createField(schema.TypeUint8, "uint8", "Uint8"),
		createField(schema.TypeUint16, "uint16", "Uint16"),
		createField(schema.TypeUint32, "uint32", "Uint32"),
		createField(schema.TypeUint64, "uint64", "Uint64"),
		createField(schema.TypeFloat32, "float32", "Float32"),
		createField(schema.TypeFloat64, "float64", "Float64"),
		createField(schema.TypeTime, "time1", "Time1"),
		createField(schema.TypeTime, "time2", "Time2"),
		createField(schema.TypeText, "text", "Text"),
	}

	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaWithComplexField(t *testing.T) {
	type Language struct {
		Name string `json:"name"`
	}

	type Post struct {
		Title string `json:"title"`
	}

	type Category struct {
		Name           string     `json:"name"`
		IgnoreSliceInt [][]int    // This field will be ignored
		SliceUint32    [][]uint32 `json:"slice_uint32" fs:"type=json"`
		SliceString    []string   `json:"slice_string" fs:"type=json"`
		Enum           string     `json:"enum" fs.enums:"[{'value':'k1','label':'Value 1'},{'value':'k2','label':'Value 2'}]"`
		IgnoreLanguage Language   `json:"ignore_language"`
		Language       Language   `json:"language" fs.relation:"{'type':'o2m','schema':'language','field':'categories'}"`
		Posts          []*Post    `json:"posts" fs.relation:"{'type':'o2m','owner':true,'optional':true,'schema':'post','field':'category'}"`
		IgnorePosts    []*Post    `ignore_json:"posts"`
	}

	ss, err := schema.CreateSchema(Category{})
	assert.NoError(t, err)
	assert.Equal(t, "category", ss.Name)
	assert.Equal(t, "categories", ss.Namespace)

	expectedFields := []*schema.Field{
		createField(schema.TypeString, "name", "Name"),
		createField(schema.TypeJSON, "slice_uint32", "Slice Uint32"),
		createField(schema.TypeJSON, "slice_string", "Slice String"),
		{
			IsSystemField: true,
			Type:          schema.TypeEnum,
			Name:          "enum",
			Label:         "Enum",
			Enums: []*schema.FieldEnum{
				{Value: "k1", Label: "Value 1"},
				{Value: "k2", Label: "Value 2"},
			},
		},
		{
			IsSystemField: true,
			Type:          schema.TypeRelation,
			Name:          "language",
			Label:         "Language",
			Relation: &schema.Relation{
				TargetSchemaName: "language",
				TargetFieldName:  "categories",
				Type:             schema.O2M,
				Owner:            false,
			},
		},
		{
			IsSystemField: true,
			Type:          schema.TypeRelation,
			Name:          "posts",
			Label:         "Posts",
			Relation: &schema.Relation{
				TargetSchemaName: "post",
				TargetFieldName:  "category",
				Type:             schema.O2M,
				Owner:            true,
				Optional:         true,
			},
		},
	}

	assert.Equal(t, expectedFields, ss.Fields)
}

func TestCreateSchemaCustomizeSchemaUseTag(t *testing.T) {
	// Case 1: Error invalid db format
	type Post struct {
		_    any    `json:"-" fs.db:"invalid"`
		Name string `json:"name"`
	}
	ss, err := schema.CreateSchema(Post{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "invalid db format")

	// Case 2: Success
	type Category struct {
		_    any    `json:"-" fs:"name=cat;namespace=cats;label_field=slug;disable_timestamp;is_junction_schema" fs.db:"{'indexes':[{'name': 'idx_name_slug','unique':true,'columns':['name','slug']}]}"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	ss, err = schema.CreateSchema(Category{})
	assert.NoError(t, err)
	assert.Equal(t, "cat", ss.Name)
	assert.Equal(t, "cats", ss.Namespace)
	assert.Equal(t, "slug", ss.LabelFieldName)
	assert.True(t, ss.DisableTimestamp)
	assert.True(t, ss.IsJunctionSchema)
	assert.Equal(t, &schema.SchemaDB{
		Indexes: []*schema.SchemaDBIndex{
			{
				Name:    "idx_name_slug",
				Unique:  true,
				Columns: []string{"name", "slug"},
			},
		},
	}, ss.DB)
}

type CustomizeMethod struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (c CustomizeMethod) Schema() *schema.Schema {
	return &schema.Schema{
		Name:             "cat",
		Namespace:        "cats",
		LabelFieldName:   "slug",
		DisableTimestamp: true,
		IsJunctionSchema: true,
		DB: &schema.SchemaDB{
			Indexes: []*schema.SchemaDBIndex{
				{
					Name:    "idx_name_slug",
					Unique:  true,
					Columns: []string{"name", "slug"},
				},
			},
		},
		Fields: []*schema.Field{
			{
				Name: "name",
				Size: 255,
			},
		},
	}
}

func TestCreateSchemaCustomizeSchemaUseMethod(t *testing.T) {
	ss, err := schema.CreateSchema(CustomizeMethod{})
	assert.NoError(t, err)

	assert.Equal(t, "cat", ss.Name)
	assert.Equal(t, "cats", ss.Namespace)
	assert.Equal(t, "slug", ss.LabelFieldName)
	assert.True(t, ss.DisableTimestamp)
	assert.True(t, ss.IsJunctionSchema)
	assert.Equal(t, int64(255), ss.Fields[0].Size)
	assert.Equal(t, &schema.SchemaDB{
		Indexes: []*schema.SchemaDBIndex{
			{
				Name:    "idx_name_slug",
				Unique:  true,
				Columns: []string{"name", "slug"},
			},
		},
	}, ss.DB)
}

type CustomizeMethodError struct {
	Name string `json:"name"`
}

func (c CustomizeMethodError) Schema() *schema.Schema {
	return &schema.Schema{
		Fields: []*schema.Field{
			{
				Name: "name",
				Size: 255,
			},
			{
				Name: "slug",
				Size: 255,
			},
		},
	}
}

func TestCreateSchemaCustomizeSchemaUseMethodError(t *testing.T) {
	ss, err := schema.CreateSchema(CustomizeMethodError{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "customized field slug not found in struct")
}

func TestCreateSchemaDuplicatedFields(t *testing.T) {
	type Category struct {
		Name  string `json:"name"`
		Slug  string `json:"slug"`
		Slug1 string `json:"slug"`
	}

	ss, err := schema.CreateSchema(Category{})
	assert.Nil(t, ss)
	assert.Contains(t, err.Error(), "field category.slug already exists")
}
