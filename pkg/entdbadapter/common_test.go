package entdbadapter

import (
	"testing"

	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestCreateEntColumn(t *testing.T) {
	type args struct {
		name   string
		field  *schema.Field
		column *entSchema.Column
	}

	tests := []args{
		{
			name: "testIDColumn",
			field: &schema.Field{
				Name:   "id",
				Type:   schema.TypeUint64,
				Unique: true,
				DB: &schema.FieldDB{
					Increment: true,
				},
			},
			column: &entSchema.Column{
				Name:      "id",
				Type:      field.TypeUint64,
				Increment: true,
				Unique:    true,
			},
		},
		{
			name: "testTextColumn",
			field: &schema.Field{
				Name: "content",
				Type: schema.TypeText,
				Size: 100,
				DB: &schema.FieldDB{
					Collation: "utf8mb4_unicode_ci",
					Key:       "MUL",
					Attr:      "UNIQUE",
				},
			},
			column: &entSchema.Column{
				Name:      "content",
				Type:      field.TypeString,
				Size:      100,
				Collation: "utf8mb4_unicode_ci",
				Key:       "MUL",
				Attr:      "UNIQUE",
			},
		},
		{
			name: "testNormalColumn",
			field: &schema.Field{
				Name:     "name",
				Type:     schema.TypeString,
				Default:  "test",
				Optional: true,
			},
			column: &entSchema.Column{
				Name:     "name",
				Type:     field.TypeString,
				Default:  "test",
				Nullable: true,
			},
		},
		{
			name: "testEnumColumn",
			field: &schema.Field{
				Name: "status",
				Type: schema.TypeEnum,
				Enums: []*schema.FieldEnum{
					{
						Label: "Active",
						Value: "active",
					},
					{
						Label: "Inactive",
						Value: "inactive",
					},
				},
			},
			column: &entSchema.Column{
				Name:  "status",
				Type:  field.TypeEnum,
				Enums: []string{"active", "inactive"},
			},
		},
		{
			name: "testTimeColumn",
			field: &schema.Field{
				Name: "created_at",
				Type: schema.TypeTime,
			},
			column: &entSchema.Column{
				Name: "created_at",
				Type: field.TypeTime,
				SchemaType: map[string]string{
					"mysql": "datetime",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := CreateEntColumn(tt.field)
			assert.Equal(t, tt.column, column)
		})
	}
}
