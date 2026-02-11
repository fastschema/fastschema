package entdbadapter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

var _ db.Client = (*Adapter)(nil)
var _ EntAdapter = (*Adapter)(nil)

// Adapter is the ent adapter for app.Client
type Adapter struct {
	config        *db.Config
	sqldb         *sql.DB
	migrationDir  string
	driver        dialect.Driver
	schemaBuilder *schema.Builder
	models        []*Model
	tables        []*entSchema.Table
	edgeSpec      map[string]sqlgraph.EdgeSpec
	typesModels   map[reflect.Type]*Model
}

func (d *Adapter) SetSQLDB(db *sql.DB) {
	d.sqldb = db
}

func (d *Adapter) SetDriver(driver dialect.Driver) {
	d.driver = driver
}

func (d *Adapter) DB() *sql.DB {
	return d.sqldb
}

func (d *Adapter) Hooks() *db.Hooks {
	if d.config.Hooks != nil {
		return d.config.Hooks()
	}

	return &db.Hooks{}
}

func (d *Adapter) Config() *db.Config {
	return d.config
}

// Dialect returns the dialect name.
func (d *Adapter) Dialect() string {
	return d.driver.Dialect()
}

// Driver returns the underlying driver.
func (d *Adapter) Driver() dialect.Driver {
	return d.driver
}

// Rollback rollbacks the transaction.
func (d *Adapter) Rollback() error {
	return nil
}

// Commit commits the transaction.
func (d *Adapter) Commit() error {
	return nil
}

// Exec executes the query and bind the values to bindValue.
func (d *Adapter) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	option := &db.QueryOption{Query: query, Args: args}
	if err := runPreDBExecHooks(ctx, d, option); err != nil {
		return nil, err
	}

	result, err := driverExec(d.driver, ctx, query, args)
	if err != nil {
		return nil, err
	}

	return result, runPostDBExecHooks(ctx, d, option, result)
}

// Query executes the query and bind the values to bindValue.
func (d *Adapter) Query(
	ctx context.Context,
	query string,
	args ...any,
) ([]*entity.Entity, error) {
	option := &db.QueryOption{Query: query, Args: args}
	if err := runPreDBQueryHooks(ctx, d, option); err != nil {
		return nil, err
	}

	entities, err := driverQuery(d.driver, ctx, option.Query, option.Args)
	if err != nil {
		return nil, err
	}

	return runPostDBQueryHooks(ctx, d, option, entities)
}

// Close closes the underlying driver.
func (d *Adapter) Close() error {
	return d.driver.Close()
}

// IsTx returns true if the client is a transaction.
func (d *Adapter) IsTx() bool {
	return false
}

// Tx creates a new transaction.
func (d *Adapter) Tx(ctx context.Context) (db.Client, error) {
	return NewTx(ctx, d)
}

// SchemaBuilder returns the schema builder.
func (d *Adapter) SchemaBuilder() *schema.Builder {
	return d.schemaBuilder
}

// Model return the model from given name.
//
//	Support finding model from name or types
//	If the input model is a string, it will use the name to find the model
//	Others, it will use the types of the input to find the model
func (d *Adapter) Model(model any) (db.Model, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	if name, ok := model.(string); ok {
		return d.model(name)
	}

	var tt reflect.Type
	if rType, ok := model.(reflect.Type); ok {
		tt = rType
	} else {
		tt = utils.GetDereferencedType(model)
	}

	typeModel, ok := d.typesModels[tt]
	if !ok {
		return nil, fmt.Errorf("model %s not found", tt.Name())
	}

	return typeModel, nil
}

func (d *Adapter) model(name string) (*Model, error) {
	for _, model := range d.models {
		if model.name == name {
			return model, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", name)
}

func (d *Adapter) init() error {
	for _, s := range d.schemaBuilder.Schemas() {
		if !s.IsJunctionSchema {
			model := d.CreateModel(s)
			d.models = append(d.models, model)
			d.tables = append(d.tables, model.entTable)
		}
	}

	for _, r := range d.schemaBuilder.Relations() {
		onDelete := relationOnDeleteOption(r)
		onUpdate := relationOnUpdateOption(r)
		currentSchema, err := d.schemaBuilder.Schema(r.SourceSchemaName)
		if err != nil {
			return err
		}

		currentModel, err := d.model(r.SourceSchemaName)
		if err != nil {
			return err
		}

		targetModel, err := d.model(r.TargetSchemaName)
		if err != nil {
			return fmt.Errorf(
				"relation model %s not found: %w",
				r.TargetSchemaName, err,
			)
		}

		targetEntColumn, err := d.resolveRelationTargetColumn(targetModel, r)
		if err != nil {
			return err
		}

		if r.FKFields != nil {
			currentModel.entTable.ForeignKeys = append(
				currentModel.entTable.ForeignKeys,
				&entSchema.ForeignKey{
					Symbol:     fmt.Sprintf("%s_%s", currentSchema.Name, r.SourceColumn),
					Columns:    []*entSchema.Column{createEntColumn(r.FKFields[0])},
					RefColumns: []*entSchema.Column{targetEntColumn},
					OnDelete:   onDelete,
					OnUpdate:   onUpdate,
					RefTable:   targetModel.entTable,
				},
			)
		}

		if r.Type == schema.M2M {
			junctionSchema := r.JunctionSchema
			junctionModel := d.CreateModel(junctionSchema, r)

			// prevent creating duplicated junction model and table
			existedModels := utils.Filter(d.models, func(m *Model) bool {
				return m.name == junctionModel.name
			})

			if len(existedModels) == 0 {
				d.models = append(d.models, junctionModel)
				d.tables = append(d.tables, junctionModel.entTable)
			}
		}
	}

	for _, r := range d.schemaBuilder.Relations() {
		relationModel, err := d.model(r.TargetSchemaName)
		if err != nil {
			return fmt.Errorf("invalid relation model %s: %w", r.TargetSchemaName, err)
		}

		targetEntColumn, err := d.resolveRelationTargetColumn(relationModel, r)
		if err != nil {
			return err
		}

		inverse := !r.IsBidi() && !r.Owner
		sourceColumn := utils.If(
			!r.Owner || r.Type.IsM2M(),
			r.SourceColumn,
			utils.If(r.BackRef != nil, r.BackRef.SourceColumn, ""),
		)
		targetColumn := utils.If(
			!r.Owner || r.Type.IsM2M(),
			r.TargetColumn,
			utils.If(r.BackRef != nil, r.BackRef.TargetColumn, ""),
		)

		// If the relation is not M2M, there should be only one foreign key column:
		//   - SourceColumn: the source schema FK column that references the target schema's PK
		// If the relation is M2M, there should be two foreign key columns in the junction table:
		//   - TargetColumn: the FK column that references the target schema's PK
		//   - SourceColumn: the FK column that references the current schema's PK
		// The order of columns depends on whether the relation is inverse.
		columns := utils.If(
			!r.Type.IsM2M(),
			[]string{sourceColumn},
			utils.If(inverse, []string{
				sourceColumn,
				targetColumn,
			}, []string{
				targetColumn,
				sourceColumn,
			}),
		)

		// If the relation is bidi and M2M, the first column is the relation field name
		// Check the implementation at: schema/builder.go:CreateM2mJunctionSchema
		// firstFKName := utils.If(r.IsBidi(), r.SchemaName, r.FieldName)
		if r.IsBidi() && r.Type.IsM2M() {
			columns[1] = r.SourceSchemaName
		}

		var edgeTargetIDSpec *sqlgraph.FieldSpec
		if r.Owner || r.IsSameType() {
			edgeTargetIDSpec = &sqlgraph.FieldSpec{
				Column: relationModel.entPrimaryColumn.Name,
				Type:   relationModel.entPrimaryColumn.Type,
			}
		} else if !r.Type.IsM2M() && targetEntColumn != nil && targetEntColumn.Name != relationModel.entPrimaryColumn.Name {
			edgeTargetIDSpec = &sqlgraph.FieldSpec{
				Column: targetEntColumn.Name,
				Type:   targetEntColumn.Type,
			}
		}

		relEdgeSpec := sqlgraph.EdgeSpec{
			Rel:     RelMaps[r.Type],
			Bidi:    r.IsBidi(),
			Inverse: inverse,
			Table:   utils.If(r.Type.IsM2M(), r.JunctionTable, relationModel.entTable.Name),
			Columns: columns,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: edgeTargetIDSpec,
			},
		}

		if r.Type.IsO2M() {
			relEdgeSpec.Rel = utils.If(r.Owner, sqlgraph.O2M, sqlgraph.M2O)
		}

		d.edgeSpec[r.Name] = relEdgeSpec
	}

	return nil
}

// NewEdgeSpec create a new edge spec
func (d *Adapter) NewEdgeSpec(
	r *schema.Relation,
	nodeIDs []driver.Value,
) (*sqlgraph.EdgeSpec, error) {
	edgeSpec, ok := d.edgeSpec[r.Name]
	if !ok {
		return nil, fmt.Errorf("invalid edgeSpec %s", r.Name)
	}

	return &sqlgraph.EdgeSpec{
		Rel:     edgeSpec.Rel,
		Bidi:    edgeSpec.Bidi,
		Inverse: edgeSpec.Inverse,
		Table:   edgeSpec.Table,
		Columns: edgeSpec.Columns,
		Target: &sqlgraph.EdgeTarget{
			IDSpec: edgeSpec.Target.IDSpec,
			Nodes:  nodeIDs,
		},
	}, nil
}

// NewEdgeStepOption creates a new edge step option
func (d *Adapter) NewEdgeStepOption(r *schema.Relation) (sqlgraph.StepOption, error) {
	edgeSpec, ok := d.edgeSpec[r.Name]
	if !ok {
		return nil, fmt.Errorf("invalid edgeSpecOption %s", r.Name)
	}

	return sqlgraph.Edge(
		edgeSpec.Rel,
		edgeSpec.Inverse,
		edgeSpec.Table,
		edgeSpec.Columns...,
	), nil
}

// CreateModel create a new model from schema
func (d *Adapter) CreateModel(s *schema.Schema, relations ...*schema.Relation) *Model {
	m := &Model{
		client:  d,
		name:    s.Name,
		schema:  s,
		columns: []*Column{},
		entTable: &entSchema.Table{
			Name:        s.Namespace,
			Columns:     []*entSchema.Column{},
			PrimaryKey:  []*entSchema.Column{},
			ForeignKeys: []*entSchema.ForeignKey{},
			Indexes:     []*entSchema.Index{},
			Annotation: &entsql.Annotation{
				Charset:   "utf8mb4",
				Collation: "utf8mb4_unicode_ci",
			},
		},
	}

	var entPrimaryColumn *entSchema.Column
	primaryFieldName := s.PrimaryKeyName()
	if primaryFieldName == "" {
		primaryFieldName = entity.FieldID
	}
	for _, f := range s.Fields {
		column := &Column{field: f}
		if !f.Type.IsRelationType() {
			entColumn := createEntColumn(f)
			m.entTable.Columns = append(m.entTable.Columns, entColumn)
			column.entColumn = entColumn

			if f.Name == primaryFieldName {
				entPrimaryColumn = entColumn
			}
		}

		m.columns = append(m.columns, column)
	}

	if entPrimaryColumn == nil {
		if len(m.entTable.Columns) > 0 {
			entPrimaryColumn = m.entTable.Columns[0]
		} else {
			entPrimaryColumn = &entSchema.Column{
				Attr:      "UNSIGNED",
				Key:       entSchema.UniqueKey,
				Type:      field.TypeUint64,
				Name:      primaryFieldName,
				Increment: true,
				Unique:    true,
			}
		}
	}

	m.entPrimaryColumn = entPrimaryColumn

	if !s.IsJunctionSchema {
		m.entTable.PrimaryKey = []*entSchema.Column{entPrimaryColumn}
	}

	// add indexes
	if s.DB != nil && s.DB.Indexes != nil {
		for _, index := range s.DB.Indexes {
			columns := make([]*entSchema.Column, len(index.Columns))
			for i, colName := range index.Columns {
				if entColumn, ok := m.entTable.Column(colName); ok {
					columns[i] = entColumn
				}
			}

			m.entTable.Indexes = append(m.entTable.Indexes, &entSchema.Index{
				Name:    index.Name,
				Unique:  index.Unique,
				Columns: columns,
			})
		}
	}

	// update junction model
	if s.IsJunctionSchema {
		if len(relations) == 0 {
			// return error because junction schema must have relation
			return m
		}

		r := relations[0]
		// add unique key for the junction table
		indexParts := []string{
			m.entTable.Columns[0].Name,
			m.entTable.Columns[1].Name,
		}
		sort.Strings(indexParts)

		colIndexes := [2]uint{0, 1}
		if indexParts[0] != m.entTable.Columns[0].Name {
			colIndexes = [2]uint{1, 0}
		}

		m.entTable.Indexes = append(m.entTable.Indexes, &entSchema.Index{
			Name:    fmt.Sprintf("unique_%s_%s", indexParts[0], indexParts[1]),
			Unique:  true,
			Columns: []*entSchema.Column{m.entTable.Columns[colIndexes[0]], m.entTable.Columns[colIndexes[1]]},
		})

		firstRelationModel := d.CreateModel(r.RelationSchemas[colIndexes[0]])
		secondRelationModel := d.CreateModel(r.RelationSchemas[colIndexes[1]])
		col1 := m.entTable.Columns[colIndexes[0]]
		col2 := m.entTable.Columns[colIndexes[1]]

		m.entTable.ForeignKeys = []*entSchema.ForeignKey{
			{
				Symbol:     fmt.Sprintf("%s_%s", s.Namespace, col1.Name),
				Columns:    []*entSchema.Column{col1},
				RefColumns: []*entSchema.Column{firstRelationModel.entPrimaryColumn},
				RefTable:   firstRelationModel.entTable,
				OnDelete:   entSchema.Cascade,
				OnUpdate:   entSchema.Cascade,
			},
			{
				Symbol:     fmt.Sprintf("%s_%s", s.Namespace, col2.Name),
				Columns:    []*entSchema.Column{col2},
				RefColumns: []*entSchema.Column{secondRelationModel.entPrimaryColumn},
				RefTable:   secondRelationModel.entTable,
				OnDelete:   entSchema.Cascade,
				OnUpdate:   entSchema.Cascade,
			},
		}
	}

	return m
}

func relationOnDeleteOption(r *schema.Relation) entSchema.ReferenceOption {
	option := r.OnDeleteOption()
	if !option.Valid() {
		return entSchema.ReferenceOption("")
	}

	return referenceOptionTypeToEnt(option)
}

func (d *Adapter) resolveRelationTargetColumn(targetModel *Model, r *schema.Relation) (*entSchema.Column, error) {
	if r.Type.IsM2M() || r.TargetColumn == "" || r.TargetColumn == targetModel.entPrimaryColumn.Name {
		return targetModel.entPrimaryColumn, nil
	}

	entColumn, ok := targetModel.entTable.Column(r.TargetColumn)
	if !ok {
		return nil, fmt.Errorf(
			"relation %s.%s: target column '%s' not found in schema %s",
			r.SourceSchemaName,
			r.SourceFieldName,
			r.TargetColumn,
			r.TargetSchemaName,
		)
	}

	return entColumn, nil
}

func relationOnUpdateOption(r *schema.Relation) entSchema.ReferenceOption {
	option := r.OnUpdateOption()
	if !option.Valid() {
		return entSchema.ReferenceOption("")
	}

	return referenceOptionTypeToEnt(option)
}

func referenceOptionTypeToEnt(option schema.ReferenceOptionType) entSchema.ReferenceOption {
	switch option {
	case schema.NoAction:
		return entSchema.NoAction
	case schema.Restrict:
		return entSchema.Restrict
	case schema.Cascade:
		return entSchema.Cascade
	case schema.SetNull:
		return entSchema.SetNull
	case schema.SetDefault:
		return entSchema.SetDefault
	default:
		return entSchema.NoAction
	}
}
