package schema

import "maps"

type SchemaDBIndex struct {
	Name    string   `json:"name,omitempty"`
	Unique  bool     `json:"unique,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

type SchemaDB struct {
	Indexes []*SchemaDBIndex `json:"indexes,omitempty"`
}

// SchemaFormZoneField defines a field in a form zone with renderer settings
type SchemaFormZoneField struct {
	Field    string         `json:"field"`              // field name
	Renderer string         `json:"renderer,omitempty"` // renderer class name
	Options  map[string]any `json:"options,omitempty"`  // renderer options/settings
}

// Clone returns a deep copy of SchemaFormZoneField
func (f *SchemaFormZoneField) Clone() *SchemaFormZoneField {
	if f == nil {
		return nil
	}
	clone := &SchemaFormZoneField{
		Field:    f.Field,
		Renderer: f.Renderer,
	}
	if f.Options != nil {
		clone.Options = make(map[string]any, len(f.Options))
		maps.Copy(clone.Options, f.Options)
	}
	return clone
}

type SchemaFormZone = []*SchemaFormZoneField

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

// Clone returns a deep copy of SchemaDBIndex
func (s *SchemaDBIndex) Clone() *SchemaDBIndex {
	if s == nil {
		return nil
	}
	columns := make([]string, len(s.Columns))
	copy(columns, s.Columns)
	return &SchemaDBIndex{
		Name:    s.Name,
		Unique:  s.Unique,
		Columns: columns,
	}
}

// Clone returns a deep copy of SchemaDB
func (s *SchemaDB) Clone() *SchemaDB {
	if s == nil {
		return nil
	}
	clone := &SchemaDB{}
	if s.Indexes != nil {
		clone.Indexes = make([]*SchemaDBIndex, len(s.Indexes))
		for i, idx := range s.Indexes {
			clone.Indexes[i] = idx.Clone()
		}
	}
	return clone
}

// Clone returns a deep copy of SchemaFormSettings
func (s *SchemaFormSettings) Clone() *SchemaFormSettings {
	if s == nil {
		return nil
	}
	clone := &SchemaFormSettings{
		ActiveView: s.ActiveView,
	}
	if s.HiddenFields != nil {
		clone.HiddenFields = make([]string, len(s.HiddenFields))
		copy(clone.HiddenFields, s.HiddenFields)
	}
	if s.Views != nil {
		clone.Views = make(map[string]SchemaFormView, len(s.Views))
		for k, v := range s.Views {
			viewClone := make(SchemaFormView, len(v))
			for zk, zv := range v {
				zoneCopy := make(SchemaFormZone, len(zv))
				for i, field := range zv {
					zoneCopy[i] = field.Clone()
				}
				viewClone[zk] = zoneCopy
			}
			clone.Views[k] = viewClone
		}
	}
	return clone
}

// Clone returns a deep copy of SchemaListSettings
func (s *SchemaListSettings) Clone() *SchemaListSettings {
	if s == nil {
		return nil
	}
	clone := &SchemaListSettings{
		ActiveView: s.ActiveView,
	}
	if s.Views != nil {
		clone.Views = make(map[string]SchemaListView, len(s.Views))
		for k, v := range s.Views {
			viewClone := make(SchemaListView, len(v))
			for i, field := range v {
				viewClone[i] = SchemaListViewField{
					Field:      field.Field,
					Invisible:  field.Invisible,
					Sortable:   field.Sortable,
					Filterable: field.Filterable,
					MaxWidth:   field.MaxWidth,
				}
			}
			clone.Views[k] = viewClone
		}
	}
	return clone
}

// Clone returns a deep copy of SchemaSettings
func (s *SchemaSettings) Clone() *SchemaSettings {
	if s == nil {
		return nil
	}
	return &SchemaSettings{
		Form: s.Form.Clone(),
		List: s.List.Clone(),
	}
}
