package entdbadapter

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// edgeConfig describes how to load an edge based on relation.
type edgeConfig struct {
	whereColumn    string        // Column on edge table to filter by (e.g., edge.FK or edge.PK)
	parentRefField *schema.Field // Field on parent to get reference values from
	isArray        bool          // True for O2M/M2M, false for O2O
}

// edgeLoader handles loading edges for a specific field.
type edgeLoader struct {
	q           *Query
	ctx         context.Context
	field       *schema.Field
	edgeModel   *Model
	edgeColumns []string // nil means select all columns
	relOpt      *db.RelationOption
}

// newEdgeLoader creates a new edge loader for the given field.
func (q *Query) newEdgeLoader(
	ctx context.Context,
	field *schema.Field,
	edgeModel *Model,
	edgeColumns []string,
	relOpt *db.RelationOption,
) *edgeLoader {
	return &edgeLoader{
		q:           q,
		ctx:         ctx,
		field:       field,
		edgeModel:   edgeModel,
		edgeColumns: edgeColumns,
		relOpt:      relOpt,
	}
}

// load executes the edge loading based on relation type:
// build config -> query neighbors -> assign to parents.
func (e *edgeLoader) load() error {
	rel := e.field.Relation
	isArrayField := rel.Type == schema.M2M || (rel.Type == schema.O2M && rel.Owner)
	if isArrayField {
		for _, node := range e.q.entities {
			if node.Get(e.field.Name) == nil {
				node.Set(e.field.Name, []*entity.Entity{})
			}
		}
	}

	if rel.Type == schema.M2M {
		return e.loadM2M()
	}

	return e.loadDirectEdge()
}

// loadDirectEdge loads O2O and O2M edges
func (e *edgeLoader) loadDirectEdge() error {
	cfg, err := e.buildDirectEdgeConfig()
	if err != nil {
		return err
	}

	// Step 1: Collect parent references (FK or PK values)
	parentRefs, parentMap, err := collectParentRefs(
		e.q.entities,
		cfg.parentRefField.Name,
		cfg.parentRefField,
		e.q.model.name,
		!e.field.Relation.Owner, // skipNullFK
	)
	if err != nil {
		return err
	}
	if len(parentRefs) == 0 {
		return nil
	}

	// Step 2: Query neighbors
	neighbors, err := e.queryNeighbors(cfg, parentRefs)
	if err != nil {
		return err
	}

	// Step 3: Assign neighbors to parents
	return e.assignNeighbors(cfg, neighbors, parentMap)
}

// buildDirectEdgeConfig (O2O,O2M) determines query parameters from relation.
func (e *edgeLoader) buildDirectEdgeConfig() (*edgeConfig, error) {
	rel := e.field.Relation
	cfg := &edgeConfig{
		// O2M owner side is an array (user.pets = many pets)
		// O2M non-owner side is single (pet.owner = one owner, essentially M2O)
		isArray: rel.Type == schema.O2M && rel.Owner,
	}

	if rel.Owner {
		// Owner side: edge has FK pointing to parent
		// Query: SELECT * FROM edge WHERE edge.FK IN (parent.PK values)
		cfg.whereColumn = rel.BackRef.SourceColumn
		targetColumn := rel.BackRef.TargetColumn

		// Determine parent reference field
		useCustomColumn := targetColumn != "" &&
			targetColumn != e.q.model.entPrimaryColumn.Name
		if useCustomColumn {
			cfg.parentRefField = e.q.model.schema.Field(targetColumn)
			if cfg.parentRefField == nil {
				return nil, fmt.Errorf("field %s.%s not found", e.q.model.name, targetColumn)
			}
		} else {
			cfg.parentRefField = e.q.model.schema.PrimaryField()
		}
	} else {
		// Non-owner side: parent has FK pointing to edge
		// Query: SELECT * FROM edge WHERE edge.PK IN (parent.FK values)
		cfg.whereColumn = rel.TargetColumn
		if cfg.whereColumn == "" {
			cfg.whereColumn = e.edgeModel.entPrimaryColumn.Name
		}

		// Parent ref field is the FK field on parent schema
		cfg.parentRefField = e.q.model.schema.Field(rel.SourceColumn)
		if cfg.parentRefField == nil {
			return nil, fmt.Errorf("field %s.%s not found", e.q.model.name, rel.SourceColumn)
		}
	}

	if cfg.parentRefField == nil {
		return nil, fmt.Errorf("schema %s is missing an id field definition", e.q.model.name)
	}

	return cfg, nil
}

// queryNeighbors executes the query to find neighbor entities.
func (e *edgeLoader) queryNeighbors(cfg *edgeConfig, parentRefs []any) ([]*entity.Entity, error) {
	// Build required columns
	requiredColumns := []string{e.edgeModel.entPrimaryColumn.Name}
	if cfg.whereColumn != e.edgeModel.entPrimaryColumn.Name {
		requiredColumns = append(requiredColumns, cfg.whereColumn)
	}

	// Build and execute query
	entEdgeQuery, err := e.buildEdgeQuery(cfg.whereColumn, parentRefs, requiredColumns)
	if err != nil {
		return nil, err
	}

	return entEdgeQuery.Get(e.ctx)
}

// assignNeighbors maps neighbors back to their parent entities.
func (e *edgeLoader) assignNeighbors(cfg *edgeConfig, neighbors []*entity.Entity, parentMap map[string][]*entity.Entity) error {
	rel := e.field.Relation
	edgeSchemaName := rel.TargetSchemaName

	for _, neighbor := range neighbors {
		// Get the reference value from neighbor to match with parent
		var refValue any
		if rel.Owner {
			// Owner: match neighbor.FK with parent.PK
			refValue = neighbor.Get(cfg.whereColumn)
		} else {
			// Non-owner: match neighbor.PK (or custom column) with parent.FK
			if cfg.whereColumn == e.edgeModel.entPrimaryColumn.Name {
				refValue = neighbor.ID()
			} else {
				refValue = neighbor.Get(cfg.whereColumn)
			}
		}

		if isZeroValue(refValue) {
			return invalidFKError(edgeSchemaName, cfg.whereColumn, neighbor.ID(), fmt.Errorf("empty reference value"))
		}

		normalized, err := normalizeIDValue(cfg.parentRefField, refValue)
		if err != nil {
			return err
		}

		key := valueKey(normalized)
		parents, ok := parentMap[key]
		if !ok {
			fkColumn := utils.If(rel.Owner, cfg.whereColumn, rel.SourceColumn)
			return noFKNodeError(e.q.model.name, edgeSchemaName, fkColumn, neighbor.ID(), refValue)
		}

		// Assign neighbor to all matching parents
		for _, parent := range parents {
			if err := e.assignToParent(parent, neighbor, cfg.isArray); err != nil {
				return err
			}
		}
	}

	return nil
}

// assignToParent assigns a single neighbor to a parent entity.
func (e *edgeLoader) assignToParent(parent, neighbor *entity.Entity, isArray bool) error {
	if !isArray {
		parent.Set(e.field.Name, neighbor)
		return nil
	}

	existing := parent.Get(e.field.Name)
	if existing == nil {
		parent.Set(e.field.Name, []*entity.Entity{neighbor})
		return nil
	}

	entities, ok := existing.([]*entity.Entity)
	if !ok {
		return invalidEntityArrayError(e.q.model.name, e.field.Name, existing)
	}

	parent.Set(e.field.Name, append(entities, neighbor))
	return nil
}

// buildEdgeQuery creates and configures an edge query with columns and relation options.
func (e *edgeLoader) buildEdgeQuery(whereColumn string, whereValues []any, requiredColumns []string) (*Query, error) {
	selectFullEdge := e.edgeColumns == nil
	colResult, err := buildEdgeColumns(e.edgeModel, e.edgeColumns, selectFullEdge, requiredColumns)
	if err != nil {
		return nil, err
	}

	edgeQuery := e.edgeModel.Query()
	// When selectFullEdge is true, directColumns is nil meaning select all (SELECT *)
	// When false, directColumns contains specific columns to select
	if len(colResult.directColumns) > 0 {
		edgeQuery = edgeQuery.Select(colResult.directColumns...)
	}
	edgeQuery = edgeQuery.Where(db.In(whereColumn, whereValues))

	entEdgeQuery, ok := edgeQuery.(*Query)
	if !ok {
		return nil, fmt.Errorf("unexpected edge query type %T", edgeQuery)
	}

	// Add nested fields and relation fields for recursive processing
	entEdgeQuery.fields = append(entEdgeQuery.fields, colResult.nestedFields...)
	entEdgeQuery.fields = append(entEdgeQuery.fields, colResult.relationFields...)

	// Apply relation options (sort, filter, select, nested options)
	if err := e.applyRelationOptions(entEdgeQuery); err != nil {
		return nil, err
	}

	// Configure per-parent limit/offset using window functions
	if e.needsPerParentLimitOffset() {
		entEdgeQuery.perParentLimit = &perParentLimitConfig{
			partitionColumn: whereColumn,
			limit:           e.relOpt.Limit,
			offset:          e.relOpt.Offset,
		}
	}

	return entEdgeQuery, nil
}

// needsPerParentLimitOffset returns true if we need per-parent limit/offset.
// Only applies to array relations (O2M owner side and M2M).
// O2M non-owner side is essentially M2O (single item per parent).
func (e *edgeLoader) needsPerParentLimitOffset() bool {
	if e.relOpt == nil || (e.relOpt.Limit == 0 && e.relOpt.Offset == 0) {
		return false
	}
	rel := e.field.Relation
	// M2M and O2M owner are array relations that need per-parent limit
	return rel.Type == schema.M2M || (rel.Type == schema.O2M && rel.Owner)
}

// =============================================================================
// M2M Edge Loading
// =============================================================================

// loadM2M loads many-to-many edges using junction table.
func (e *edgeLoader) loadM2M() error {
	// Collect parent IDs
	parentIDs, parentByID, err := collectEntityIDs(e.q.model.name, e.q.model.schema.PrimaryField(), e.q.entities)
	if err != nil {
		return err
	}

	// Use window function approach if limit/offset is specified
	if e.needsPerParentLimitOffset() {
		return e.loadM2MWithWindowFunction(parentIDs, parentByID)
	}

	// neighborParents maps neighbor ID → set of parent entities
	neighborParents := make(map[string]map[*entity.Entity]struct{})

	// Build and execute M2M query with junction table
	entEdgeQuery, err := e.buildM2MQuery(parentIDs, parentByID, neighborParents)
	if err != nil {
		return err
	}

	// M2M default ordering by primary key
	entEdgeQuery.order = []string{e.edgeModel.entPrimaryColumn.Name}

	// Apply relation options (may override default order)
	if err := e.applyRelationOptions(entEdgeQuery); err != nil {
		return err
	}

	// Execute query
	neighbors, err := entEdgeQuery.Get(e.ctx)
	if err != nil {
		return err
	}

	// Assign neighbors to parents (no limit/offset filtering needed here)
	return e.assignM2MNeighborsSimple(neighbors, neighborParents)
}

// loadM2MWithWindowFunction loads M2M edges using window functions for per-parent limit/offset.
func (e *edgeLoader) loadM2MWithWindowFunction(parentIDs []driver.Value, parentByID map[string]*entity.Entity) error {
	colCfg := e.getM2MColumnConfig()

	// Get edge columns
	colResult, _ := buildEdgeColumns(e.edgeModel, e.edgeColumns, false, nil)
	cols := colResult.directColumns
	if len(cols) == 0 || len(colResult.relationFields) > 0 || len(colResult.nestedFields) > 0 {
		cols = e.edgeModel.DBColumns()
	}
	if !utils.Contains(cols, e.edgeModel.entPrimaryColumn.Name) {
		cols = append([]string{e.edgeModel.entPrimaryColumn.Name}, cols...)
	}

	entAdapter, ok := e.q.client.(EntAdapter)
	if !ok {
		return fmt.Errorf("client is not an ent adapter")
	}

	// Build the query with window function
	entities, junctionValues, err := e.executeM2MWindowQuery(entAdapter, parentIDs, colCfg, cols)
	if err != nil {
		return err
	}

	// Map neighbors to parents using junction values
	for i, neighbor := range entities {
		if i >= len(junctionValues) {
			break
		}

		junctionValue := junctionValues[i]
		parent, ok := parentByID[valueKey(junctionValue)]
		if !ok {
			continue
		}

		// Append neighbor to parent's edge array
		existing := parent.Get(e.field.Name).([]*entity.Entity)
		parent.Set(e.field.Name, append(existing, neighbor))
	}

	return nil
}

// executeM2MWindowQuery builds and executes the M2M query with window functions.
func (e *edgeLoader) executeM2MWindowQuery(
	entAdapter EntAdapter,
	parentIDs []driver.Value,
	colCfg *m2mColumnConfig,
	cols []string,
) ([]*entity.Entity, []any, error) {
	rel := e.field.Relation
	builder := sql.Dialect(entAdapter.Driver().Dialect())

	edgeTable := builder.Table(e.edgeModel.schema.Namespace)
	junction := builder.Table(rel.JunctionTable)

	// Build ORDER BY for window function
	orderCols := []string{e.edgeModel.entPrimaryColumn.Name}
	if e.relOpt != nil && e.relOpt.Sort != "" {
		orderCols = []string{e.relOpt.Sort}
	}

	// Build inner query: SELECT junction.parent_id, edge.*, ROW_NUMBER() OVER (...) AS row_num
	// FROM edge JOIN junction ON junction.edge_id = edge.id WHERE junction.parent_id IN (...)
	inner := builder.Select().
		From(edgeTable).
		Join(junction).On(junction.C(colCfg.joinColumn), edgeTable.C(e.edgeModel.entPrimaryColumn.Name))

	// Add junction parent column (for mapping back to parents)
	inner.AppendSelect(junction.C(colCfg.selectColumn) + " AS _junction_parent_id")

	// Build window function - use unqualified column name since junction table gets aliased
	// We partition by the selectColumn (parent reference), not the joinColumn
	windowFn := sql.RowNumber().PartitionBy(colCfg.selectColumn)
	for _, col := range orderCols {
		if after, ok := strings.CutPrefix(col, "-"); ok {
			windowFn = windowFn.OrderBy(edgeTable.C(after) + " DESC")
		} else {
			windowFn = windowFn.OrderBy(edgeTable.C(col))
		}
	}

	// Add edge columns
	for _, col := range cols {
		inner.AppendSelect(edgeTable.C(col))
	}

	// Add window function
	inner.AppendSelectExprAs(windowFn, "row_num")

	// Add WHERE clause for parent IDs
	inner.Where(sql.InValues(junction.C(colCfg.conditionColumn), parentIDs...))

	// Apply filter predicate if specified
	if e.relOpt != nil && e.relOpt.Filter != nil {
		schemaBuilder := e.q.client.SchemaBuilder()
		if schemaBuilder != nil {
			predicates, err := db.CreatePredicatesFromRelationFilter(schemaBuilder, e.edgeModel.schema, e.relOpt.Filter)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid relation filter for %s: %w", e.field.Name, err)
			}
			sqlPredicatesFn, err := createEntPredicates(entAdapter, e.edgeModel, predicates)
			if err != nil {
				return nil, nil, err
			}
			inner.Where(sql.And(sqlPredicatesFn(inner)...))
		}
	}

	// Alias the inner query
	inner.As("ranked")

	// Build outer query with row_num filter
	outerCols := append([]string{"_junction_parent_id"}, cols...)
	outer := builder.Select(outerCols...).From(inner)

	// Apply row_num conditions for per-parent limit/offset
	if e.relOpt.Offset > 0 {
		outer.Where(sql.GT("row_num", e.relOpt.Offset))
	}
	if e.relOpt.Limit > 0 {
		maxRowNum := e.relOpt.Offset + e.relOpt.Limit
		outer.Where(sql.LTE("row_num", maxRowNum))
	}

	// Apply ordering to outer query
	for _, col := range orderCols {
		if after, ok := strings.CutPrefix(col, "-"); ok {
			outer.OrderBy(sql.Desc(after))
		} else {
			outer.OrderBy(sql.Asc(col))
		}
	}

	// Execute query
	query, args := outer.Query()
	return e.executeM2MWindowQueryRaw(entAdapter, query, args, cols)
}

// executeM2MWindowQueryRaw executes the raw M2M window query and parses results.
func (e *edgeLoader) executeM2MWindowQueryRaw(
	entAdapter EntAdapter,
	query string,
	args []any,
	cols []string,
) ([]*entity.Entity, []any, error) {
	var rows = &sql.Rows{}
	if err := entAdapter.Driver().Query(e.ctx, query, args, rows); err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var entities []*entity.Entity
	var junctionValues []any

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	for rows.Next() {
		// Create scan destinations: first column is junction parent ID, rest are edge columns
		values := make([]any, len(columns))

		// Junction parent ID scanner (first column)
		rel := e.field.Relation
		junctionSchema := rel.JunctionSchema
		colCfg := e.getM2MColumnConfig()
		selectField := junctionSchema.Field(colCfg.selectColumn)
		values[0] = columnScanValue(selectField.Type)

		// Edge column scanners
		for i := 1; i < len(columns); i++ {
			colName := columns[i]
			field := e.edgeModel.schema.Field(colName)
			if field != nil {
				values[i] = columnScanValue(field.Type)
			} else {
				values[i] = new(any)
			}
		}

		if err := rows.Scan(values...); err != nil {
			return nil, nil, err
		}

		// Extract junction value
		junctionValue, err := columnAssignValue(colCfg.selectColumn, selectField.Type, values[0], entity.New())
		if err != nil {
			return nil, nil, err
		}
		junctionValues = append(junctionValues, junctionValue)

		// Build entity from remaining columns
		ent := entity.New()
		for i := 1; i < len(columns); i++ {
			colName := columns[i]
			field := e.edgeModel.schema.Field(colName)
			if field != nil {
				val, err := columnAssignValue(colName, field.Type, values[i], ent)
				if err != nil {
					return nil, nil, err
				}
				ent.Set(colName, val)
			}
		}
		entities = append(entities, ent)
	}

	return entities, junctionValues, nil
}

// applyRelationOptions applies relation options (sort, filter, select, nested options) to an edge query.
func (e *edgeLoader) applyRelationOptions(q *Query) error {
	if e.relOpt == nil {
		return nil
	}

	// Apply sort order
	if e.relOpt.Sort != "" {
		q.order = []string{e.relOpt.Sort}
	}

	// Apply filter
	if e.relOpt.Filter != nil {
		builder := e.q.client.SchemaBuilder()
		if builder != nil {
			predicates, err := db.CreatePredicatesFromRelationFilter(builder, e.edgeModel.schema, e.relOpt.Filter)
			if err != nil {
				return fmt.Errorf("invalid relation filter for %s: %w", e.field.Name, err)
			}
			q.predicates = append(q.predicates, predicates...)
		}
	}

	// Apply select fields
	if e.relOpt.Select != nil {
		for _, sel := range e.relOpt.Select {
			if !utils.Contains(q.fields, sel) {
				q.fields = append(q.fields, sel)
			}
		}
	}

	// Pass nested relation options
	q.relationOptions = e.q.relationOptions.GetNestedOptions(e.field.Name)

	return nil
}

// m2mColumnConfig holds the junction table column configuration.
type m2mColumnConfig struct {
	conditionColumn string // Junction column to filter by (WHERE clause)
	joinColumn      string // Junction column that joins to edge PK
	selectColumn    string // Junction column to select (for mapping back to parent)
}

// getM2MColumnConfig determines the junction table columns based on relation metadata.
func (e *edgeLoader) getM2MColumnConfig() *m2mColumnConfig {
	rel := e.field.Relation

	cfg := &m2mColumnConfig{}

	// Condition column: filter junction rows by parent IDs
	cfg.conditionColumn = utils.If(rel.IsBidi(), rel.SourceSchemaName, rel.BackRef.SourceFieldName)
	if !rel.IsBidi() && rel.TargetColumn != "" {
		cfg.conditionColumn = rel.TargetColumn
	}

	// Join column: join junction to edge table
	cfg.joinColumn = utils.If(rel.IsBidi(), rel.SourceSchemaName, rel.SourceFieldName)
	if rel.SourceColumn != "" {
		cfg.joinColumn = rel.SourceColumn
	}

	// Select column: map neighbors back to parents
	cfg.selectColumn = rel.BackRef.SourceFieldName
	if !rel.IsBidi() && rel.TargetColumn != "" {
		cfg.selectColumn = rel.TargetColumn
	}

	return cfg
}

// buildM2MQuery creates the M2M query with junction table join.
func (e *edgeLoader) buildM2MQuery(
	parentIDs []driver.Value,
	parentByID map[string]*entity.Entity,
	neighborParents map[string]map[*entity.Entity]struct{},
) (*Query, error) {
	rel := e.field.Relation
	colCfg := e.getM2MColumnConfig()

	// Get edge columns configuration
	colResult, _ := buildEdgeColumns(e.edgeModel, e.edgeColumns, false, nil)

	// Create base query
	edgeQuery := e.edgeModel.Query()
	entEdgeQuery, ok := edgeQuery.(*Query)
	if !ok {
		return nil, fmt.Errorf("unexpected edge query type %T", edgeQuery)
	}

	// Add nested and relation fields for recursive processing
	entEdgeQuery.fields = append(entEdgeQuery.fields, colResult.nestedFields...)
	entEdgeQuery.fields = append(entEdgeQuery.fields, colResult.relationFields...)

	// Build M2M predicate with junction table join
	entEdgeQuery.querySpec.Predicate = e.buildM2MPredicate(parentIDs, colCfg, colResult)

	// Get junction schema field for scan/assign
	junctionSchema := rel.JunctionSchema
	if junctionSchema == nil {
		return nil, fmt.Errorf("relation %s.%s missing junction schema", rel.SourceSchemaName, rel.SourceFieldName)
	}
	selectField := junctionSchema.Field(colCfg.selectColumn)
	if selectField == nil {
		return nil, fmt.Errorf("junction column %s not found for relation %s.%s", colCfg.selectColumn, rel.SourceSchemaName, rel.SourceFieldName)
	}

	// Setup custom scan/assign to capture junction data
	e.setupM2MScanAssign(entEdgeQuery, colCfg.selectColumn, selectField, parentByID, neighborParents)

	return entEdgeQuery, nil
}

// buildM2MPredicate creates the predicate function for M2M queries.
func (e *edgeLoader) buildM2MPredicate(parentIDs []driver.Value, colCfg *m2mColumnConfig, colResult *edgeColumnResult) func(s *sql.Selector) {
	rel := e.field.Relation

	return func(s *sql.Selector) {
		// Join with junction table
		junction := sql.Table(rel.JunctionTable)
		s.Join(junction).On(junction.C(colCfg.joinColumn), s.C(e.edgeModel.entPrimaryColumn.Name))

		// Filter by parent IDs
		s.Where(sql.InValues(junction.C(colCfg.conditionColumn), parentIDs...))

		// Select junction column for mapping (aliased to avoid conflicts)
		s.Select(junction.C(colCfg.selectColumn) + " AS " + colCfg.selectColumn + "_id")

		// Select edge columns
		cols := colResult.directColumns
		if len(cols) == 0 || len(colResult.relationFields) > 0 || len(colResult.nestedFields) > 0 {
			cols = e.edgeModel.DBColumns()
		}
		if !utils.Contains(cols, e.edgeModel.entPrimaryColumn.Name) {
			cols = append([]string{e.edgeModel.entPrimaryColumn.Name}, cols...)
		}

		s.AppendSelect(utils.Map(cols, func(c string) string {
			return s.C(c)
		})...)
		s.SetDistinct(false)
	}
}

// setupM2MScanAssign configures custom scan/assign functions to capture junction data.
func (e *edgeLoader) setupM2MScanAssign(
	q *Query,
	selectColumn string,
	selectField *schema.Field,
	parentByID map[string]*entity.Entity,
	neighborParents map[string]map[*entity.Entity]struct{},
) {
	originalAssign := q.querySpec.Assign
	originalScan := q.querySpec.ScanValues

	// Custom ScanValues: prepend junction column scanner
	q.querySpec.ScanValues = func(columns []string) ([]any, error) {
		values, err := originalScan(columns[1:])
		if err != nil {
			return nil, err
		}
		junctionScanner := columnScanValue(selectField.Type)
		return append([]any{junctionScanner}, values...), nil
	}

	// Custom Assign: capture parent-neighbor relationship
	q.querySpec.Assign = func(columns []string, values []any) error {
		// Extract junction value (parent ID)
		junctionValue, err := columnAssignValue(selectColumn, selectField.Type, values[0], entity.New())
		if err != nil {
			return err
		}

		// Assign remaining values to entity
		if err := originalAssign(columns[1:], values[1:]); err != nil {
			return err
		}

		if junctionValue == nil {
			return fmt.Errorf("junction column %s returned nil", selectColumn)
		}

		// Find parent entity
		parent, ok := parentByID[valueKey(junctionValue)]
		if !ok {
			return fmt.Errorf("no base entity found for junction value %v", junctionValue)
		}

		// Get the neighbor that was just assigned
		if len(q.entities) == 0 {
			return fmt.Errorf("edge assignment missing neighbor entity for %v", junctionValue)
		}
		neighbor := q.entities[len(q.entities)-1]

		// Record parent-neighbor relationship
		neighborKey := valueKey(neighbor.ID())
		if neighborParents[neighborKey] == nil {
			neighborParents[neighborKey] = make(map[*entity.Entity]struct{})
		}
		neighborParents[neighborKey][parent] = struct{}{}

		return nil
	}
}

// assignM2MNeighborsSimple assigns M2M neighbors to parent entities without limit/offset filtering.
// This is used when no limit/offset is specified, or when window functions handle the filtering.
func (e *edgeLoader) assignM2MNeighborsSimple(
	neighbors []*entity.Entity,
	neighborParents map[string]map[*entity.Entity]struct{},
) error {
	for _, neighbor := range neighbors {
		neighborKey := valueKey(neighbor.ID())
		parents, ok := neighborParents[neighborKey]
		if !ok {
			continue
		}

		for parent := range parents {
			// Append neighbor to parent's edge array
			existing := parent.Get(e.field.Name).([]*entity.Entity)
			parent.Set(e.field.Name, append(existing, neighbor))
		}

		// Clean up processed neighbor
		delete(neighborParents, neighborKey)
	}

	return nil
}

// =============================================================================
// Public API
// =============================================================================

// loadEdges loads the edges for the given edge selections.
func (q *Query) loadEdges(ctx context.Context, edges map[string]*edgeSelection) error {
	for _, edge := range edges {
		relation := edge.field.Relation
		edgeModel, err := q.client.Model(relation.TargetSchemaName)
		if err != nil {
			return err
		}

		edgeEntModel, ok := edgeModel.(*Model)
		if !ok {
			return fmt.Errorf("unexpected model type %T", edgeModel)
		}

		relOpt := q.relationOptions.Get(edge.field.Name)

		loader := q.newEdgeLoader(ctx, edge.field, edgeEntModel, edge.columns, relOpt)
		if err := loader.load(); err != nil {
			return err
		}
	}
	return nil
}

// edgeColumnResult holds the result of buildEdgeColumns.
type edgeColumnResult struct {
	directColumns  []string
	nestedFields   []string
	relationFields []string
}

// buildEdgeColumns separates edge columns into direct columns, nested field paths, and relation fields.
// It also ensures required columns (like primary key and FK columns) are included.
func buildEdgeColumns(
	edgeModel *Model,
	edgeColumns []string,
	selectFullEdge bool,
	requiredColumns []string,
) (*edgeColumnResult, error) {
	result := &edgeColumnResult{
		directColumns:  []string{},
		nestedFields:   []string{},
		relationFields: []string{},
	}

	for _, col := range edgeColumns {
		if strings.Contains(col, ".") {
			result.nestedFields = append(result.nestedFields, col)
			continue
		}

		column, err := edgeModel.Column(col)
		if err != nil {
			return nil, fmt.Errorf("invalid column %q for model %s: %w", col, edgeModel.name, err)
		}

		if column.field.Type.IsRelationType() {
			result.relationFields = append(result.relationFields, col)
		} else {
			result.directColumns = append(result.directColumns, col)
		}
	}

	// If selectFullEdge is true, we want all columns (nil means select all)
	if selectFullEdge {
		result.directColumns = nil
	} else {
		// Ensure required columns are included
		for _, reqCol := range requiredColumns {
			if reqCol != "" && !utils.Contains(result.directColumns, reqCol) {
				result.directColumns = append(result.directColumns, reqCol)
			}
		}
		result.directColumns = utils.Unique(result.directColumns)
	}

	return result, nil
}
