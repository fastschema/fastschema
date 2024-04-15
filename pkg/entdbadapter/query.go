package entdbadapter

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
)

type Query struct {
	limit           uint
	offset          uint
	fields          []string
	order           []string
	entities        []*schema.Entity
	predicates      []*app.Predicate
	withEdgesFields []*schema.Field
	client          app.DBClient
	model           *Model
	querySpec       *sqlgraph.QuerySpec
}

func (q *Query) Options() *app.QueryOption {
	return &app.QueryOption{
		Limit:      q.limit,
		Offset:     q.offset,
		Columns:    q.fields,
		Order:      q.order,
		Predicates: q.predicates,
		Model:      q.model,
	}
}

// Limit sets the limit of the query.
func (q *Query) Limit(limit uint) app.Query {
	q.limit = limit
	return q
}

// Offset sets the offset of the query.
func (q *Query) Offset(offset uint) app.Query {
	q.offset = offset
	return q
}

// Order sets the order of the query.
func (q *Query) Order(order ...string) app.Query {
	q.order = append(q.order, order...)
	return q
}

// Select sets the columns of the query.
func (q *Query) Select(fields ...string) app.Query {
	q.fields = append(q.fields, fields...)
	return q
}

// Where adds the given predicates to the query.
func (q *Query) Where(predicates ...*app.Predicate) app.Query {
	q.predicates = append(q.predicates, predicates...)
	return q
}

// Count returns the number of entities that match the query.
func (q *Query) Count(options *app.CountOption, ctxs ...context.Context) (int, error) {
	ctxs = append(ctxs, context.Background())
	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return 0, fmt.Errorf("client is not an ent adapter")
	}

	if options != nil {
		q.querySpec.Unique = options.Unique
		if options.Column != "" {
			q.querySpec.Node.Columns = []string{options.Column}
		}
	}

	if len(q.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
		if err != nil {
			return 0, err
		}
		q.querySpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	return sqlgraph.CountNodes(ctxs[0], entAdapter.Driver(), q.querySpec)
}

// First returns the first entity that matches the query.
func (q *Query) First(ctxs ...context.Context) (*schema.Entity, error) {
	q.Limit(1)
	entities, err := q.Get(ctxs...)

	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, &app.NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

// Only returns the matched entity or an error if there is more than one.
func (q *Query) Only(ctxs ...context.Context) (*schema.Entity, error) {
	entities, err := q.Get(ctxs...)

	if err != nil {
		return nil, err
	}

	if len(entities) > 1 {
		return nil, fmt.Errorf("more than one entity found")
	}

	if len(entities) == 0 {
		return nil, &app.NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

// Get returns the list of entities that match the query.
func (q *Query) Get(ctxs ...context.Context) ([]*schema.Entity, error) {
	ctxs = append(ctxs, context.Background())
	columnNames := []string{}
	edgeColumns := map[string][]string{}
	allSelectsAreEdges := true

	if len(q.fields) > 0 {
		columnNames = append(columnNames, q.model.entIDColumn.Name)

		for _, columnName := range q.fields {
			if strings.Contains(columnName, ".") {
				parts := strings.Split(columnName, ".")
				if len(parts) != 2 {
					return nil, fmt.Errorf(`invalid column name %q`, columnName)
				}

				columnName = parts[0]
				edgeColumns[columnName] = append(edgeColumns[columnName], parts[1])
			}

			column, err := q.model.Column(columnName)
			if err != nil {
				return nil, err
			}

			if column.field.Type.IsRelationType() {
				existedEdgeFields := utils.Filter(q.withEdgesFields, func(f *schema.Field) bool {
					return f.Name == column.field.Name
				})

				if len(existedEdgeFields) > 0 {
					continue
				}

				relation := column.field.Relation
				q.withEdgesFields = append(q.withEdgesFields, column.field)
				if relation.Type != schema.M2M && !relation.Owner {
					columnNames = append(columnNames, relation.GetTargetFKColumn())
				}
				continue
			}

			if columnName != q.model.entIDColumn.Name {
				columnNames = append(columnNames, columnName)
				allSelectsAreEdges = false
			}
		}
	}

	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return nil, fmt.Errorf("client is not an ent adapter")
	}

	builder := sql.Dialect(entAdapter.Driver().Dialect())
	if !allSelectsAreEdges {
		q.querySpec.Node.Columns = columnNames
	}
	q.querySpec.From = builder.
		Select(columnNames...).
		From(builder.Table(q.model.schema.Namespace))

	if len(q.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
		if err != nil {
			return nil, err
		}
		q.querySpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	if len(q.order) > 0 {
		orderSelectors := []func(*sql.Selector){}

		for _, order := range q.order {
			orderFn := sql.Asc
			columnName := order

			if strings.HasPrefix(order, "-") {
				columnName = strings.TrimPrefix(order, "-")
				orderFn = sql.Desc
			}

			column, err := q.model.Column(columnName)
			if err != nil {
				return nil, err
			}

			if !column.field.Sortable {
				return nil, fmt.Errorf(`column %q is not sortable`, columnName)
			}

			orderSelectors = append(orderSelectors, func(s *sql.Selector) {
				s.OrderBy(orderFn(columnName))
			})
		}

		q.querySpec.Order = func(s *sql.Selector) {
			for _, orderSelector := range orderSelectors {
				orderSelector(s)
			}
		}
	}

	if q.limit > 0 {
		q.querySpec.Limit = int(q.limit)
	}

	if q.offset > 0 {
		q.querySpec.Offset = int(q.offset)
	}

	if err := sqlgraph.QueryNodes(ctxs[0], entAdapter.Driver(), q.querySpec); err != nil {
		return nil, err
	}

	if err := q.loadEdges(edgeColumns); err != nil {
		return nil, err
	}

	var hooks = &app.Hooks{}
	if q.client != nil {
		hooks = q.client.Hooks()
	}

	if len(hooks.PostDBGet) > 0 {
		queryOptions := q.Options()
		for _, hook := range hooks.PostDBGet {
			var err error
			if q.entities, err = hook(queryOptions, q.entities); err != nil {
				return nil, err
			}
		}
	}

	return q.entities, nil
}

// loadEdges loads the edges for the given fields.
func (q *Query) loadEdges(edgesColumns map[string][]string) error {
	for _, edgeField := range q.withEdgesFields {
		relation := edgeField.Relation
		edgeModel, err := q.client.Model(relation.TargetSchemaName)
		if err != nil {
			return err
		}

		edgeEntModel, ok := edgeModel.(*Model)
		if !ok {
			return fmt.Errorf(`unexpected model type %T`, edgeModel)
		}

		edgeColumns := edgesColumns[edgeField.Name]

		if relation.Type == schema.O2M {
			if err := q.loadEdgesO2M(edgeField, edgeEntModel, edgeColumns); err != nil {
				return err
			}
		}

		if relation.Type == schema.O2O {
			if err := q.loadEdgesO2O(edgeField, edgeEntModel, edgeColumns); err != nil {
				return err
			}
		}

		if relation.Type == schema.M2M {
			if err := q.loadEdgesM2M(edgeField, edgeEntModel, edgeColumns); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadEdgesO2M loads the one-to-many edges for the given field.
func (q *Query) loadEdgesO2M(field *schema.Field, edgeModel *Model, edgeColumns []string) error {
	if !field.Relation.Owner {
		return q.loadNonOwnerEdges(field, edgeModel, edgeColumns)
	}

	return q.loadOwnerEdges(field, edgeModel, edgeColumns, func(node *schema.Entity, neighbor *schema.Entity) error {
		edgeValues := node.Get(field.Name)
		if edgeValues == nil {
			node.Set(field.Name, []*schema.Entity{neighbor})
			return nil
		}

		edgeEntities, ok := edgeValues.([]*schema.Entity)
		if !ok {
			return invalidEntityArrayError(q.model.name, field.Name, edgeValues)
		}

		edgeEntities = append(edgeEntities, neighbor)
		node.Set(field.Name, edgeEntities)
		return nil
	})
}

// loadEdgesO2O loads the one-to-one edges for the given field.
func (q *Query) loadEdgesO2O(field *schema.Field, edgeModel *Model, edgeColumns []string) error {
	if !field.Relation.Owner {
		return q.loadNonOwnerEdges(field, edgeModel, edgeColumns)
	}

	return q.loadOwnerEdges(field, edgeModel, edgeColumns, func(node *schema.Entity, neighbor *schema.Entity) error {
		node.Set(field.Name, neighbor)
		return nil
	})
}

// loadEdgesM2M loads the many-to-many edges for the given field.
func (q *Query) loadEdgesM2M(field *schema.Field, edgeModel *Model, edgeColumns []string) error {
	edgeIDs := make([]driver.Value, len(q.entities))
	byID := make(map[uint64]*schema.Entity)
	nids := make(map[uint64]map[*schema.Entity]struct{})
	for i, node := range q.entities {
		edgeIDs[i] = node.ID()
		byID[node.ID()] = node
		node.Set(field.Name, make([]*schema.Entity, 0))
	}

	edgeQuery := edgeModel.Query()
	entEdgeQuery, ok := edgeQuery.(*Query)
	if !ok {
		return fmt.Errorf(`unexpected edge query type %T`, edgeQuery)
	}

	relation := field.Relation
	conditionColumn := utils.If(relation.IsBidi(), relation.SchemaName, relation.BackRef.FieldName)
	joinColumn := utils.If(relation.IsBidi(), relation.SchemaName, relation.FieldName)
	entEdgeQuery.querySpec.Predicate = func(s *sql.Selector) {
		joinJuntion := sql.Table(relation.JunctionTable)
		s.Join(joinJuntion).On(joinJuntion.C(joinColumn), s.C(edgeModel.entIDColumn.Name))
		s.Where(sql.InValues(joinJuntion.C(conditionColumn), edgeIDs...))
		s.Select(joinJuntion.C(relation.BackRef.FieldName) + " AS " + relation.BackRef.FieldName + "_id")

		if len(edgeColumns) == 0 {
			edgeColumns = edgeModel.DBColumns()
		}
		if !utils.Contains(edgeColumns, edgeModel.entIDColumn.Name) {
			edgeColumns = append([]string{edgeModel.entIDColumn.Name}, edgeColumns...)
		}

		s.AppendSelect(utils.Map(edgeColumns, func(c string) string {
			return s.C(c)
		})...)
		s.SetDistinct(false)
	}

	assignFn := entEdgeQuery.querySpec.Assign
	valuesFn := entEdgeQuery.querySpec.ScanValues
	entEdgeQuery.querySpec.ScanValues = func(columns []string) ([]any, error) {
		values, err := valuesFn(columns[1:])
		if err != nil {
			return nil, err
		}
		return append([]any{new(sql.NullInt64)}, values...), nil
	}

	entEdgeQuery.querySpec.Assign = func(columns []string, values []any) error {
		outValue := uint64(values[0].(*sql.NullInt64).Int64)
		inValue := uint64(values[1].(*sql.NullInt64).Int64)
		if nids[inValue] == nil {
			nids[inValue] = map[*schema.Entity]struct{}{byID[outValue]: {}}
			return assignFn(columns[1:], values[1:])
		}
		nids[inValue][byID[outValue]] = struct{}{}
		return nil
	}

	entEdgeQuery.order = []string{edgeModel.entIDColumn.Name}
	neighbors, err := entEdgeQuery.Get(q.model.ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		nodes, ok := nids[n.ID()]
		if !ok {
			continue
		}
		for kn := range nodes {
			kn.Set(field.Name, append(kn.Get(field.Name).([]*schema.Entity), n))
		}
	}

	return nil
}

// loadNonOwnerEdges loads the non-owner edges for the given field.
func (q *Query) loadNonOwnerEdges(field *schema.Field, edgeModel *Model, edgeColumns []string) error {
	relation := field.Relation
	edgeSchemaName := relation.TargetSchemaName
	ids := make([]any, 0, len(q.entities))
	nodeids := make(map[uint64][]*schema.Entity)
	fkColumn := relation.GetTargetFKColumn()

	for _, entity := range q.entities {
		fkUint64, err := entity.GetUint64(fkColumn, true)
		if err != nil {
			return invalidFKError(edgeSchemaName, fkColumn, entity.ID(), err)
		}

		if fkUint64 == 0 {
			continue
		}

		if _, ok := nodeids[fkUint64]; !ok {
			ids = append(ids, fkUint64)
		}
		nodeids[fkUint64] = append(nodeids[fkUint64], entity)
	}

	if len(ids) == 0 {
		return nil
	}

	edgeQuery := edgeModel.Query().Select(edgeColumns...).Where(app.In(edgeModel.entIDColumn.Name, ids))
	neighbors, err := edgeQuery.Get(q.model.ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID()]
		if !ok {
			return noFKNodeError(q.model.name, edgeSchemaName, fkColumn, n.ID(), 0)
		}

		for i := range nodes {
			nodes[i].Set(field.Name, n)
		}
	}

	return nil
}

// loadOwnerEdges loads the owner edges for the given field.
func (q *Query) loadOwnerEdges(
	field *schema.Field,
	edgeModel *Model,
	edgeColumns []string,
	assignFn func(node, neighbor *schema.Entity) error,
) error {
	relation := field.Relation
	edgeSchemaName := relation.TargetSchemaName
	fks := make([]any, 0, len(q.entities))
	nodeids := make(map[uint64]*schema.Entity)
	fkColumn := relation.BackRef.GetTargetFKColumn()

	for _, entity := range q.entities {
		entityID := entity.ID()
		fks = append(fks, entityID)
		nodeids[entityID] = entity
	}

	if len(edgeColumns) > 0 && !utils.Contains(edgeColumns, fkColumn) {
		edgeColumns = append(edgeColumns, fkColumn)
	}

	neighbors, err := edgeModel.Query().Select(edgeColumns...).Where(app.In(fkColumn, fks)).Get(q.model.ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		fkUint64, err := n.GetUint64(fkColumn, false)
		if err != nil {
			return invalidFKError(edgeSchemaName, fkColumn, n.ID(), err)
		}

		node, ok := nodeids[fkUint64]
		if !ok {
			return noFKNodeError(q.model.name, edgeSchemaName, fkColumn, n.ID(), fkUint64)
		}

		if err := assignFn(node, n); err != nil {
			return err
		}
	}

	return nil
}

// scanValues create a slice of scan values for the given columns.
func scanValues(s *schema.Schema, columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		field, err := s.Field(columns[i])
		if err != nil { // This error is not found error. Ignore it.
			values[i] = new(any)
			continue
		}
		switch field.Type {
		case schema.TypeJSON, schema.TypeBytes:
			values[i] = new([]byte)
		case schema.TypeBool:
			values[i] = new(sql.NullBool)
		case schema.TypeFloat32, schema.TypeFloat64:
			values[i] = new(sql.NullFloat64)
		case schema.TypeInt8, schema.TypeInt16, schema.TypeInt32, schema.TypeInt, schema.TypeInt64, schema.TypeUint8, schema.TypeUint16, schema.TypeUint32, schema.TypeUint, schema.TypeUint64:
			values[i] = new(sql.NullInt64)
		case schema.TypeEnum, schema.TypeString, schema.TypeText:
			values[i] = new(sql.NullString)
		case schema.TypeTime:
			values[i] = new(sql.NullTime)
		case schema.TypeUUID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for schema %s", columns[i], s.Name)
		}
	}
	return values, nil
}

// assignValues assigns the given values to the entity.
func assignValues(s *schema.Schema, entity *schema.Entity, columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		field, err := s.Field(columns[i])
		if err != nil { // This error is not found error. Ignore it.
			entity.Set(columns[i], new(any))
			continue
		}
		switch field.Type {
		case schema.TypeBool:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fieldTypeError("Bool", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Bool)
			}
		case schema.TypeTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fieldTypeError("Time", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Time)
			}
		case schema.TypeJSON:
			if value, ok := values[i].(*[]byte); !ok {
				return fieldTypeError("JSON", values[i])
			} else if value != nil && len(*value) > 0 {
				e := entity.Get(field.Name)
				if err := json.Unmarshal(*value, &e); err != nil {
					return fmt.Errorf("unmarshal field field_type_JSON: %w", err)
				}
				entity.Set(field.Name, e)
			}
		case schema.TypeUUID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fieldTypeError("UUID", values[i])
			} else if value != nil {
				entity.Set(field.Name, *value)
			}
		case schema.TypeBytes:
			if value, ok := values[i].(*[]byte); !ok {
				return fieldTypeError("Bytes", values[i])
			} else if value != nil {
				entity.Set(field.Name, *value)
			}
		case schema.TypeEnum:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fieldTypeError("Enum", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.String)
			}
		case schema.TypeString:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fieldTypeError("String", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.String)
			}
		case schema.TypeText:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fieldTypeError("Text", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.String)
			}
		case schema.TypeInt8:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int8", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int8(value.Int64))
			}
		case schema.TypeInt16:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int16", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int16(value.Int64))
			}
		case schema.TypeInt32:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int32", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int32(value.Int64))
			}
		case schema.TypeInt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int", values[i])
			} else if value.Valid {
				entity.Set(field.Name, int(value.Int64))
			}
		case schema.TypeInt64:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Int64", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Int64)
			}
		case schema.TypeUint8:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint8", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint8(value.Int64))
			}
		case schema.TypeUint16:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint16", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint16(value.Int64))
			}
		case schema.TypeUint32:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint32", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint32(value.Int64))
			}
		case schema.TypeUint:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fieldTypeError("Uint", values[i])
			}

			entity.Set(field.Name, utils.If(value.Valid, uint(value.Int64), uint(0)))
		case schema.TypeUint64:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fieldTypeError("Uint64", values[i])
			} else if value.Valid {
				entity.Set(field.Name, uint64(value.Int64))
			}
		case schema.TypeFloat32:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fieldTypeError("Float32", values[i])
			} else if value.Valid {
				entity.Set(field.Name, float32(value.Float64))
			}
		case schema.TypeFloat64:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fieldTypeError("Float64", values[i])
			} else if value.Valid {
				entity.Set(field.Name, value.Float64)
			}
		}
	}

	return nil
}

func fieldTypeError(fieldType string, fieldValue any) error {
	return fmt.Errorf("unexpected type %T for field type %s", fieldValue, fieldType)
}

func invalidFKError(edgeSchemaName, fkColumn string, id uint64, err error) error {
	return fmt.Errorf(
		`invalid FK value %s.%s for node id=%d: %w`,
		edgeSchemaName, fkColumn, id, err,
	)
}

func noFKNodeError(schemaName, edgeSchemaName, fkColumn string, id, fk uint64) error {
	return fmt.Errorf(
		`no FK node (%s) found for (%s=%v).%s=%v`,
		schemaName, edgeSchemaName, id, fkColumn, fk,
	)
}

func invalidEntityArrayError(schemaName, fieldName string, edgeValues any) error {
	return fmt.Errorf(
		`edge values %s.%s=%v (%T) is not []*schema.Entity`,
		schemaName, fieldName, edgeValues, edgeValues,
	)
}
