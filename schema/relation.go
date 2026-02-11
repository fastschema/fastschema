package schema

import (
	"fmt"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
)

// Relation define the relation structure
type Relation struct {
	BackRef          *Relation `json:"-"` // back reference relation
	Name             string    `json:"-"` // relation name: auto generated
	SourceSchemaName string    `json:"-"` // the source schema name
	SourceFieldName  string    `json:"-"` // the source schema field name

	TargetSchemaName string `json:"schema"`          // the target schema name
	TargetFieldName  string `json:"field,omitempty"` // the target schema field name

	Type  RelationType `json:"type"`            // the relation type: o2o, o2m, m2m
	Owner bool         `json:"owner,omitempty"` // the relation owner: true, false
	// OnDelete specifies the action to take on delete
	// e.g., "NO ACTION", "RESTRICT", "CASCADE", "SET NULL", "SET DEFAULT".
	// Only used for O2O/O2M non-owner side
	OnDelete ReferenceOptionType `json:"on_delete,omitempty"`

	// OnUpdate specifies the action to take on update
	// e.g., "NO ACTION", "RESTRICT", "CASCADE", "SET NULL", "SET DEFAULT".
	// Only used for O2O/O2M non-owner side
	OnUpdate ReferenceOptionType `json:"on_update,omitempty"`

	// SourceColumn is the FK column name in the source schema table
	// (for O2O/O2M non-owner side)
	// or the column referencing the target schema's PK
	// in the M2M junction table.
	SourceColumn string `json:"source_column"`
	// TargetColumn optionally specifies the referenced column on the target schema
	// for FK relations. For M2M relations it continues to describe the junction
	// column that references the source schema's primary key.
	TargetColumn string `json:"target_column"`

	// JunctionTable is the junction table name for m2m relation
	JunctionTable   string    `json:"junction_table,omitempty"`
	Optional        bool      `json:"optional"`
	FKFields        []*Field  `json:"-"`
	RelationSchemas []*Schema `json:"-"` // for m2m relation
	JunctionSchema  *Schema   `json:"-"` // for m2m relation
}

// Init initialize the relation
func (r *Relation) Init(schema *Schema, relationSchema *Schema, f *Field) *Relation {
	r.Optional = f.Optional
	r.SourceFieldName = f.Name
	r.SourceSchemaName = schema.Name
	r.Name = fmt.Sprintf(
		"%s.%s-%s.%s",
		schema.Name,
		f.Name,
		r.TargetSchemaName,
		r.TargetFieldName,
	)

	if r.HasFKs() && !r.OnDelete.Valid() {
		r.OnDelete = r.defaultOnDeleteOption()
	}

	if r.HasFKs() && !r.OnUpdate.Valid() {
		r.OnUpdate = r.defaultOnUpdateOption()
	}

	if r.HasFKs() {
		r.SourceColumn = utils.If(
			r.SourceColumn == "",
			r.SourceFieldName+"_id",
			r.SourceColumn,
		)

		if !r.Type.IsM2M() {
			r.TargetColumn = utils.If(
				r.TargetColumn == "",
				entity.FieldID,
				r.TargetColumn,
			)
		}
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
		Name:             r.Name,
		SourceSchemaName: r.SourceSchemaName,
		SourceFieldName:  r.SourceFieldName,

		TargetSchemaName: r.TargetSchemaName,
		TargetFieldName:  r.TargetFieldName,

		Type:         r.Type,
		Owner:        r.Owner,
		OnDelete:     r.OnDelete,
		OnUpdate:     r.OnUpdate,
		Optional:     r.Optional,
		SourceColumn: r.SourceColumn,
		TargetColumn: r.TargetColumn,
	}

	return newRelation
}

func (r *Relation) defaultOnDeleteOption() ReferenceOptionType {
	if r.Optional {
		return SetNull
	}

	return NoAction
}

// OnDeleteOption returns the effective on delete option for relations with FKs.
func (r *Relation) OnDeleteOption() ReferenceOptionType {
	if !r.HasFKs() {
		return ReferenceOptionTypeInvalid
	}

	if r.OnDelete.Valid() {
		return r.OnDelete
	}

	return r.defaultOnDeleteOption()
}

func (r *Relation) defaultOnUpdateOption() ReferenceOptionType {
	return NoAction
}

// OnUpdateOption returns the effective on update option for relations with FKs.
func (r *Relation) OnUpdateOption() ReferenceOptionType {
	if !r.HasFKs() {
		return ReferenceOptionTypeInvalid
	}

	if r.OnUpdate.Valid() {
		return r.OnUpdate
	}

	return r.defaultOnUpdateOption()
}

// GetBackRefName get the back reference name
func (r *Relation) GetBackRefName() string {
	return fmt.Sprintf(
		"%s.%s-%s.%s",
		r.TargetSchemaName,
		r.TargetFieldName,
		r.SourceSchemaName,
		r.SourceFieldName,
	)
}

// IsSameType check if the relation is same type
func (r *Relation) IsSameType() bool {
	return r.SourceSchemaName == r.TargetSchemaName
}

// IsBidi check if the relation is bidirectional
func (r *Relation) IsBidi() bool {
	return r.IsSameType() && r.SourceFieldName == r.TargetFieldName
}

// HasFKs check if the relation has foreign keys
func (r *Relation) HasFKs() bool {
	isO2OTwoTypeNotOwner := r.Type.IsO2O() && !r.IsSameType() && !r.Owner
	isO2OSameTypeRecursiveNotOwner := r.Type.IsO2O() && r.IsSameType() && !r.IsBidi() && !r.Owner
	isO2OBidi := r.Type.IsO2O() && r.IsBidi()
	isO2mNotOwner := r.Type.IsO2M() && !r.Owner
	return isO2OTwoTypeNotOwner || isO2OSameTypeRecursiveNotOwner || isO2mNotOwner || isO2OBidi
}

// CreateFKField create the foreign key fields
func (r *Relation) CreateFKField() (*Field, error) {
	if !r.HasFKs() {
		return nil, nil
	}

	fkColumn := r.SourceColumn
	fkField := &Field{
		IsSystemField: true,
		Immutable:     true,
		Type:          TypeUint64,
		Name:          fkColumn,
		Label:         fkColumn,
		Unique:        r.Type.IsO2O(),
		Optional:      r.Optional,
		DB: &FieldDB{
			Key:  utils.If(r.Type.IsO2O(), "UNI", ""),
			Attr: "UNSIGNED",
		},
	}

	if err := fkField.Init(); err != nil {
		return nil, err
	}

	return fkField, nil
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
		relation.SourceSchemaName,
		relation.SourceFieldName,
		relation.TargetSchemaName,
		relation.TargetFieldName,
		relation.TargetSchemaName,
		relation.TargetFieldName,
	)
}
