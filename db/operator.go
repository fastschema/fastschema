package db

import (
	"bytes"
	"encoding/json"
)

// OperatorType is the type of the operator
type OperatorType int

const (
	OpInvalid OperatorType = iota
	OpEQ
	OpNEQ
	OpGT
	OpGTE
	OpLT
	OpLTE
	OpLike
	OpNotLike
	OpContains
	OpNotContains
	OpContainsFold
	OpNotContainsFold
	OpIN
	OpNIN
	OpNULL
	endOperatorTypes
)

var (
	operatorTypeToStrings = [...]string{
		OpInvalid:         "invalid",
		OpEQ:              "$eq",
		OpNEQ:             "$neq",
		OpGT:              "$gt",
		OpGTE:             "$gte",
		OpLT:              "$lt",
		OpLTE:             "$lte",
		OpLike:            "$like",
		OpNotLike:         "$notlike",
		OpContains:        "$contains",
		OpNotContains:     "$notcontains",
		OpContainsFold:    "$containsfold",
		OpNotContainsFold: "$notcontainsfold",
		OpIN:              "$in",
		OpNIN:             "$nin",
		OpNULL:            "$null",
	}

	stringToOperatorTypes = map[string]OperatorType{
		"invalid":          OpInvalid,
		"$eq":              OpEQ,
		"$neq":             OpNEQ,
		"$gt":              OpGT,
		"$gte":             OpGTE,
		"$lt":              OpLT,
		"$lte":             OpLTE,
		"$like":            OpLike,
		"$notlike":         OpNotLike,
		"$contains":        OpContains,
		"$notcontains":     OpNotContains,
		"$containsfold":    OpContainsFold,
		"$notcontainsfold": OpNotContainsFold,
		"$in":              OpIN,
		"$nin":             OpNIN,
		"$null":            OpNULL,
	}
)

// String returns the string representation of a type.
func (t OperatorType) String() string {
	if t < endOperatorTypes {
		return operatorTypeToStrings[t]
	}
	return operatorTypeToStrings[OpInvalid]
}

// Valid reports if the given type if known type.
func (t OperatorType) Valid() bool {
	return t > OpInvalid && t < endOperatorTypes
}

// MarshalJSON marshal an enum value to the quoted json string value
func (t OperatorType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(operatorTypeToStrings[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *OperatorType) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*t = stringToOperatorTypes[j] // If the string can't be found, it will be set to the zero value: 'invalid'
	return nil
}

func And(predicates ...*Predicate) *Predicate {
	return &Predicate{And: predicates}
}

func Or(predicates ...*Predicate) *Predicate {
	return &Predicate{Or: predicates}
}

// EQ creates an equality predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func EQ(field string, value any) *Predicate {
	return &Predicate{Field: field, Operator: OpEQ, Value: value}
}

// NEQ creates a not-equal predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func NEQ(field string, value any) *Predicate {
	return &Predicate{Field: field, Operator: OpNEQ, Value: value}
}

// GT creates a greater-than predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func GT(field string, value any) *Predicate {
	return &Predicate{Field: field, Operator: OpGT, Value: value}
}

// GTE creates a greater-than-or-equal predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func GTE(field string, value any) *Predicate {
	return &Predicate{Field: field, Operator: OpGTE, Value: value}
}

// LT creates a less-than predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func LT(field string, value any) *Predicate {
	return &Predicate{Field: field, Operator: OpLT, Value: value}
}

// LTE creates a less-than-or-equal predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func LTE(field string, value any) *Predicate {
	return &Predicate{Field: field, Operator: OpLTE, Value: value}
}

// Like creates a LIKE predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func Like(field string, value string) *Predicate {
	return &Predicate{Field: field, Operator: OpLike, Value: value}
}

// NotLike creates a NOT LIKE predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func NotLike(field string, value string) *Predicate {
	return &Predicate{Field: field, Operator: OpNotLike, Value: value}
}

// Contains creates a contains predicate (substring match).
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func Contains(field string, value string) *Predicate {
	return &Predicate{Field: field, Operator: OpContains, Value: value}
}

// NotContains creates a not-contains predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func NotContains(field string, value string) *Predicate {
	return &Predicate{Field: field, Operator: OpNotContains, Value: value}
}

// ContainsFold creates a case-insensitive contains predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func ContainsFold(field string, value string) *Predicate {
	return &Predicate{Field: field, Operator: OpContainsFold, Value: value}
}

// NotContainsFold creates a case-insensitive not-contains predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func NotContainsFold(field string, value string) *Predicate {
	return &Predicate{Field: field, Operator: OpNotContainsFold, Value: value}
}

// In creates an IN predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func In[T any](field string, values []T) *Predicate {
	return &Predicate{Field: field, Operator: OpIN, Value: values}
}

// NotIn creates a NOT IN predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func NotIn[T any](field string, values []T) *Predicate {
	return &Predicate{Field: field, Operator: OpNIN, Value: values}
}

// Null creates a NULL check predicate.
// The field can be a simple field name (e.g., "name") or a dot notation path
// for relation fields (e.g., "teams.slug" where "teams" is the relation field
// and "slug" is the target field in the related schema).
func Null(field string, value bool) *Predicate {
	return &Predicate{Field: field, Operator: OpNULL, Value: value}
}

// IsFalse creates a predicate that checks if a boolean field is false.
// The field can be a simple field name (e.g., "active") or a dot notation path
// for relation fields (e.g., "teams.active" where "teams" is the relation field
// and "active" is the target field in the related schema).
func IsFalse(field string) *Predicate {
	return &Predicate{Field: field, Operator: OpEQ, Value: false}
}

// IsTrue creates a predicate that checks if a boolean field is true.
// The field can be a simple field name (e.g., "active") or a dot notation path
// for relation fields (e.g., "teams.active" where "teams" is the relation field
// and "active" is the target field in the related schema).
func IsTrue(field string) *Predicate {
	return &Predicate{Field: field, Operator: OpEQ, Value: true}
}
