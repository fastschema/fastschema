package schema

type SchemaDBIndex struct {
	Name    string   `json:"name,omitempty"`
	Unique  bool     `json:"unique,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

type SchemaDB struct {
	Indexes []*SchemaDBIndex `json:"indexes,omitempty"`
}

type SchemaFormZone = []string

type SchemaFormView map[string]SchemaFormZone

type SchemaFormSettings struct {
	ActiveView   string                    `json:"active_view,omitempty"`
	HiddenFields []string                  `json:"hidden_fields,omitempty"`
	Views        map[string]SchemaFormView `json:"views,omitempty"`
}

type SchemaListViewField struct {
	Field      string `json:"field,omitempty"`
	Invisible  bool   `json:"invisible,omitempty"`
	Sortable   bool   `json:"sortable,omitempty"`
	Filterable bool   `json:"filterable,omitempty"`
	MaxWidth   int    `json:"max_width,omitempty"`
}

type SchemaListView []SchemaListViewField

type SchemaListSettings struct {
	ActiveView string                    `json:"active_view,omitempty"`
	Views      map[string]SchemaListView `json:"views,omitempty"`
}

type SchemaSettings struct {
	Form *SchemaFormSettings `json:"form,omitempty"`
	List *SchemaListSettings `json:"list,omitempty"`
}
