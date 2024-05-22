package schema

import (
	"fmt"

	"github.com/fastschema/fastschema/pkg/utils"
)

type RelationFKColumns struct {
	CurrentColumn string `json:"current_column"`
	TargetColumn  string `json:"target_column"`
}

// Reation define the relation structure
type Relation struct {
	BackRef    *Relation `json:"-"` // back reference relation
	Name       string    `json:"-"` // relation name: auto generated
	SchemaName string    `json:"-"` // schema name: get from the current schema
	FieldName  string    `json:"-"` // field name: get from the current field

	TargetSchemaName string `json:"schema"`          // target schema name
	TargetFieldName  string `json:"field,omitempty"` // target field name, aka the back reference field name

	Type            RelationType       `json:"type"`            // the relation type: o2o, o2m, m2m
	Owner           bool               `json:"owner,omitempty"` // the relation owner: true, false
	FKColumns       *RelationFKColumns `json:"fk_columns"`
	JunctionTable   string             `json:"junction_table,omitempty"` // junction table name for m2m relation
	Optional        bool               `json:"optional"`
	FKFields        []*Field           `json:"-"`
	RelationSchemas []*Schema          `json:"-"` // for m2m relation
	JunctionSchema  *Schema            `json:"-"` // for m2m relation
}

// Init initialize the relation
func (r *Relation) Init(schema *Schema, relationSchema *Schema, f *Field) *Relation {
	r.Optional = f.Optional
	r.FieldName = f.Name
	r.SchemaName = schema.Name
	r.Name = fmt.Sprintf(
		"%s.%s-%s.%s",
		schema.Name,
		f.Name,
		r.TargetSchemaName,
		r.TargetFieldName,
	)

	if r.HasFKs() {
		r.FKColumns = utils.If(
			r.FKColumns == nil,
			&RelationFKColumns{TargetColumn: f.Name + "_id"},
			r.FKColumns,
		)
	}

	return r
}

// Clone clone the relation
func (r *Relation) Clone() *Relation {
	if r == nil {
		return nil
	}

	// Skip clone auto generated fields
	newRelation := &Relation{
		Name:       r.Name,
		SchemaName: r.SchemaName,
		FieldName:  r.FieldName,

		TargetSchemaName: r.TargetSchemaName,
		TargetFieldName:  r.TargetFieldName,
		Type:             r.Type,
		Owner:            r.Owner,
		Optional:         r.Optional,
	}

	return newRelation
}

// GetBackRefName get the back reference name
func (r *Relation) GetBackRefName() string {
	return fmt.Sprintf(
		"%s.%s-%s.%s",
		r.TargetSchemaName,
		r.TargetFieldName,
		r.SchemaName,
		r.FieldName,
	)
}

// IsSameType check if the relation is same type
func (r *Relation) IsSameType() bool {
	return r.SchemaName == r.TargetSchemaName
}

// IsBidi check if the relation is bidirectional
func (r *Relation) IsBidi() bool {
	return r.IsSameType() && r.FieldName == r.TargetFieldName
}

// GetFKColumns return the foreign key columns
func (r *Relation) GetFKColumns() *RelationFKColumns {
	if r.HasFKs() {
		return r.FKColumns
	}

	return nil
}

// GetTargetFKColumn return the FK column name for o2m and o2o relation
func (r *Relation) GetTargetFKColumn() string {
	// fkKeys := utils.GetMapKeys(r.FKColumns)

	// if len(fkKeys) > 0 {
	// 	return r.FKColumns[fkKeys[0]]
	// }

	// return ""

	return r.FKColumns.TargetColumn
}

// HasFKs check if the relation has foreign keys
func (r *Relation) HasFKs() bool {
	isO2OTwoTypeNotOwner := r.Type.IsO2O() && !r.IsSameType() && !r.Owner
	isO2OSameTypeRecursiveNotOwner := r.Type.IsO2O() && r.IsSameType() && !r.IsBidi() && !r.Owner
	isO2OBidi := r.Type.IsO2O() && r.IsBidi()
	isO2mNotOwner := r.Type.IsO2M() && !r.Owner
	return isO2OTwoTypeNotOwner || isO2OSameTypeRecursiveNotOwner || isO2mNotOwner || isO2OBidi
}

// CreateFKFields create the foreign key fields
func (r *Relation) CreateFKFields() *Field {
	if !r.HasFKs() {
		return nil
	}

	fk := r.GetTargetFKColumn()
	fkField := &Field{
		IsSystemField: true,
		Type:          TypeUint64,
		Name:          fk,
		Label:         fk,
		Unique:        r.Type.IsO2O(),
		Optional:      r.Optional,
		DB: &FieldDB{
			Key:  utils.If(r.Type.IsO2O(), "UNI", ""),
			Attr: "UNSIGNED",
		},
	}

	fkField.Init()
	return fkField
}

func NewRelationNodeError(schema *Schema, field *Field) error {
	return fmt.Errorf(
		"relation node %s.%s: '%s' is not found",
		schema.Name,
		field.Name,
		field.Relation.TargetSchemaName,
	)
}

func NewRelationBackRefError(relation *Relation) error {
	return fmt.Errorf(
		"backref relation for %s.%s is not valid: '%s.%s', please check the 'field' property in the '%s.%s' relation definition",
		relation.SchemaName,
		relation.FieldName,
		relation.TargetSchemaName,
		relation.TargetFieldName,
		relation.TargetSchemaName,
		relation.TargetFieldName,
	)
}
