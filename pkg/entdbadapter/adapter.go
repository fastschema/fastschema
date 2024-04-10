package entdbadapter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sort"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// Adapter is the ent adapter for app.Client
type Adapter struct {
	config        *app.DBConfig
	sqldb         *sql.DB
	migrationDir  string
	driver        dialect.Driver
	schemaBuilder *schema.Builder
	models        []*Model
	tables        []*entSchema.Table
	edgeSpec      map[string]sqlgraph.EdgeSpec
	hooks         *app.Hooks
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

func (d *Adapter) Hooks() *app.Hooks {
	return d.hooks
}

func (d *Adapter) Config() *app.DBConfig {
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
	args,
	bindValue any,
) error {
	return d.driver.Exec(ctx, query, args, bindValue)
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
func (d *Adapter) Tx(ctx context.Context) (app.DBClient, error) {
	return NewTx(ctx, d)
}

// SchemaBuilder returns the schema builder.
func (d *Adapter) SchemaBuilder() *schema.Builder {
	return d.schemaBuilder
}

// Model return the model object for query and mutation.
func (d *Adapter) Model(name string) (app.Model, error) {
	return d.model(name)
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
		onDelete := utils.If(r.Optional, entSchema.SetNull, entSchema.NoAction)
		targetSchema, err := d.schemaBuilder.Schema(r.SchemaName)
		if err != nil {
			return err
		}

		currentModel, err := d.model(r.SchemaName)
		if err != nil {
			return err
		}

		targetModel, err := d.model(r.TargetSchemaName)
		if err != nil {
			return fmt.Errorf(
				"relation models %s or %s not found: %w",
				r.SchemaName,
				r.TargetSchemaName,
				err,
			)
		}

		if r.FKFields != nil {
			currentModel.entTable.ForeignKeys = append(
				currentModel.entTable.ForeignKeys,
				&entSchema.ForeignKey{
					Symbol:     fmt.Sprintf("%s_%s", targetSchema.Name, r.GetTargetFKColumn()),
					Columns:    []*entSchema.Column{createEntColumn(r.FKFields[0])},
					RefColumns: []*entSchema.Column{targetModel.entIDColumn},
					OnDelete:   onDelete,
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

		inverse := !r.IsBidi() && !r.Owner
		fkColumns := utils.If(
			!r.Owner || r.Type.IsM2M(),
			r.FKColumns,
			r.BackRef.FKColumns,
		)

		// If the relation is not M2M, there should be only one foreign key column.
		// the column is the foreign key column of the target schema
		// If the relation is M2M, there should be two foreign key columns.
		// the first column is the foreign key column of the current schema
		// the second column is the foreign key column of the target schema
		columns := utils.If(
			!r.Type.IsM2M(),
			[]string{fkColumns.TargetColumn},
			utils.If(inverse, []string{
				fkColumns.TargetColumn,
				fkColumns.CurrentColumn,
			}, []string{
				fkColumns.CurrentColumn,
				fkColumns.TargetColumn,
			}),
		)

		// If the relation is bidi and M2M, the first column is the relation field name
		// Check the implementation at: schema/builder.go:CreateM2mJunctionSchema
		// firstFKName := utils.If(r.IsBidi(), r.SchemaName, r.FieldName)
		if r.IsBidi() && r.Type.IsM2M() {
			columns[1] = r.SchemaName
		}

		relEdgeSpec := sqlgraph.EdgeSpec{
			Rel:     RelMaps[r.Type],
			Bidi:    r.IsBidi(),
			Inverse: inverse,
			Table:   utils.If(r.Type.IsM2M(), r.JunctionTable, relationModel.entTable.Name),
			Columns: columns,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: utils.If(r.Owner || r.IsSameType(), &sqlgraph.FieldSpec{
					Column: relationModel.entIDColumn.Name,
					Type:   relationModel.entIDColumn.Type,
				}, nil),
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

// NewEdgeStep create a new edge step
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

func (d *Adapter) CreateDBModel(s *schema.Schema, relations ...*schema.Relation) app.Model {
	return d.CreateModel(s, relations...)
}

// CreateModel create a new model from schema
func (d *Adapter) CreateModel(s *schema.Schema, relations ...*schema.Relation) *Model {
	m := &Model{
		client:  d,
		name:    s.Name,
		schema:  s,
		ctx:     context.Background(),
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

	m.entIDColumn = &entSchema.Column{
		Attr:      "UNSIGNED",
		Key:       entSchema.UniqueKey,
		Type:      field.TypeUint64,
		Name:      schema.FieldID,
		Increment: true,
		Unique:    true,
	}

	if !s.IsJunctionSchema {
		m.entTable.PrimaryKey = []*entSchema.Column{m.entIDColumn}
	}

	for _, f := range s.Fields {
		column := &Column{field: f}
		if !f.Type.IsRelationType() {
			entColumn := createEntColumn(f)
			m.entTable.Columns = append(m.entTable.Columns, entColumn)
			column.entColumn = entColumn
		}

		m.columns = append(m.columns, column)
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
		indexParts := []string{m.entTable.Columns[0].Name, m.entTable.Columns[1].Name}
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
				RefColumns: []*entSchema.Column{firstRelationModel.entIDColumn},
				RefTable:   firstRelationModel.entTable,
				OnDelete:   entSchema.Cascade,
			},
			{
				Symbol:     fmt.Sprintf("%s_%s", s.Namespace, col2.Name),
				Columns:    []*entSchema.Column{col2},
				RefColumns: []*entSchema.Column{secondRelationModel.entIDColumn},
				RefTable:   secondRelationModel.entTable,
				OnDelete:   entSchema.Cascade,
			},
		}
	}

	return m
}
