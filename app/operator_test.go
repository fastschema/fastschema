package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func jsonEqual(a, b []byte) bool {
	var aj, bj any
	if err := json.Unmarshal(a, &aj); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bj); err != nil {
		return false
	}
	return aj == bj
}

func TestOperatorTypeString(t *testing.T) {
	// Test valid types
	for i := OpEQ; i < endOperatorTypes; i++ {
		expected := operatorTypeToStrings[i]
		actual := i.String()
		if actual != expected {
			t.Errorf("String() for OperatorType %d was incorrect, got: %s, want: %s.", i, actual, expected)
		}
	}

	// Test invalid type
	invalidType := OpInvalid
	expected := operatorTypeToStrings[OpInvalid]
	actual := invalidType.String()
	if actual != expected {
		t.Errorf("String() for OperatorType %d was incorrect, got: %s, want: %s.", invalidType, actual, expected)
	}

	// Test cast from int to OperatorType invalid in String()
	invalidType = OperatorType(100)
	expected = operatorTypeToStrings[OpInvalid]
	actual = invalidType.String()
	if actual != expected {
		t.Errorf("String() for OperatorType %d was incorrect, got: %s, want: %s.", invalidType, actual, expected)
	}
}

func TestOperatorTypeValid(t *testing.T) {
	// Test valid types
	for i := OpEQ; i < endOperatorTypes; i++ {
		if !i.Valid() {
			t.Errorf("Valid() for OperatorType %d should have been true.", i)
		}
	}

	// Test invalid type
	invalidType := OpInvalid
	if invalidType.Valid() {
		t.Errorf("Valid() for OperatorType %d should have been false.", invalidType)
	}
}

func TestOperatorTypeMarshalJSON(t *testing.T) {
	// Test valid types
	for i := OpEQ; i < endOperatorTypes; i++ {
		expected, _ := json.Marshal(operatorTypeToStrings[i])
		actual, err := i.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON() for OperatorType %d threw an error: %v.", i, err)
		}
		if !jsonEqual(actual, expected) {
			t.Errorf("MarshalJSON() for OperatorType %d was incorrect, got: %v, want: %v.", i, actual, expected)
		}
	}

	// Test invalid type
	invalidType := OpInvalid
	expected, _ := json.Marshal(operatorTypeToStrings[OpInvalid])
	actual, err := invalidType.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() for OperatorType %d threw an error: %v.", invalidType, err)
	}
	if !jsonEqual(actual, expected) {
		t.Errorf("MarshalJSON() for OperatorType %d was incorrect, got: %v, want: %v.", invalidType, actual, expected)
	}
}

func TestOperatorTypeUnmarshalJSON(t *testing.T) {
	// Test valid type
	bytes, _ := json.Marshal("$eq")
	expected := OpEQ
	var actual OperatorType
	err := actual.UnmarshalJSON(bytes)
	if err != nil {
		t.Errorf("UnmarshalJSON() for OperatorType %d threw an error: %v.", OpEQ, err)
	}
	if actual != expected {
		t.Errorf("UnmarshalJSON() for OperatorType %d was incorrect, got: %v, want: %v.", OpEQ, actual, expected)
	}

	// Test invalid type
	bytes, _ = json.Marshal("invalid")
	expected = OpInvalid
	err = actual.UnmarshalJSON(bytes)
	if err != nil {
		t.Errorf("UnmarshalJSON() for OperatorType %d threw an error: %v.", OpInvalid, err)
	}
	if actual != expected {
		t.Errorf("UnmarshalJSON() for OperatorType %d was incorrect, got: %v, want: %v.", OpInvalid, actual, expected)
	}

	// Test unmarshal error
	bytes = []byte("invalid")
	err = actual.UnmarshalJSON(bytes)
	if err == nil {
		t.Errorf("UnmarshalJSON() for OperatorType %d should have thrown an error.", OpInvalid)
	}
}

func TestAndOr(t *testing.T) {
	p1 := EQ("field1", 1)
	p2 := GT("field2", 2)
	p3 := IsFalse("field3")
	p4 := LTE("field4", 4)

	ops := map[string]func(predicates ...*Predicate) *Predicate{
		"And": And,
		"Or":  Or,
	}
	for op, fn := range ops {
		// with 2 predicates
		expected := &Predicate{}

		if op == "And" {
			expected.And = []*Predicate{p1, p2}
		} else {
			expected.Or = []*Predicate{p1, p2}
		}

		actual := fn(p1, p2)
		if !predicatesEqual(actual, expected) {
			t.Errorf("%s() with 2 predicates was incorrect, got: %v, want: %v.", op, actual, expected)
		}

		// with 3 predicates
		expected = &Predicate{}
		if op == "And" {
			expected.And = []*Predicate{p1, p2, p3}
		} else {
			expected.Or = []*Predicate{p1, p2, p3}
		}

		actual = fn(p1, p2, p3)
		if !predicatesEqual(actual, expected) {
			t.Errorf("%s() with 3 predicates was incorrect, got: %v, want: %v.", op, actual, expected)
		}

		// with 4 predicates
		expected = &Predicate{}
		if op == "And" {
			expected.And = []*Predicate{p1, p2, p3, p4}
		} else {
			expected.Or = []*Predicate{p1, p2, p3, p4}
		}

		actual = fn(p1, p2, p3, p4)
		if !predicatesEqual(actual, expected) {
			t.Errorf("%s() with 4 predicates was incorrect, got: %v, want: %v.", op, actual, expected)
		}

		// with 0 predicates
		expected = &Predicate{}
		actual = fn()
		if !predicatesEqual(actual, expected) {
			t.Errorf("%s() with 0 predicates was incorrect, got: %v, want: %v.", op, actual, expected)
		}
	}
}

func TestEQ(t *testing.T) {
	// Test EQ with int value
	expected := &Predicate{
		Field:    "field1",
		Operator: OpEQ,
		Value:    1,
	}
	actual := EQ("field1", 1)
	if !predicatesEqual(actual, expected) {
		t.Errorf("EQ() with int value was incorrect, got: %v, want: %v.", actual, expected)
	}

	// Test EQ with string value
	expected = &Predicate{
		Field:    "field2",
		Operator: OpEQ,
		Value:    "value2",
	}
	actual = EQ("field2", "value2")
	if !predicatesEqual(actual, expected) {
		t.Errorf("EQ() with string value was incorrect, got: %v, want: %v.", actual, expected)
	}

	// Test EQ with bool value
	expected = &Predicate{
		Field:    "field3",
		Operator: OpEQ,
		Value:    true,
	}
	actual = EQ("field3", true)
	if !predicatesEqual(actual, expected) {
		t.Errorf("EQ() with bool value was incorrect, got: %v, want: %v.", actual, expected)
	}

	// Test EQ with nil value
	expected = &Predicate{
		Field:    "field4",
		Operator: OpEQ,
		Value:    nil,
	}
	actual = EQ("field4", nil)
	if !predicatesEqual(actual, expected) {
		t.Errorf("EQ() with nil value was incorrect, got: %v, want: %v.", actual, expected)
	}
}

func TestNEQ(t *testing.T) {
	// Test NEQ with int value
	expected := &Predicate{
		Field:    "field1",
		Operator: OpNEQ,
		Value:    1,
	}
	actual := NEQ("field1", 1)
	if !predicatesEqual(actual, expected) {
		t.Errorf("NEQ() with int value was incorrect, got: %v, want: %v.", actual, expected)
	}

	// Test NEQ with string value
	expected = &Predicate{
		Field:    "field2",
		Operator: OpNEQ,
		Value:    "value",
	}
	actual = NEQ("field2", "value")
	if !predicatesEqual(actual, expected) {
		t.Errorf("NEQ() with string value was incorrect, got: %v, want: %v.", actual, expected)
	}

	// Test NEQ with nil value
	expected = &Predicate{
		Field:    "field3",
		Operator: OpNEQ,
		Value:    nil,
	}
	actual = NEQ("field3", nil)
	if !predicatesEqual(actual, expected) {
		t.Errorf("NEQ() with nil value was incorrect, got: %v, want: %v.", actual, expected)
	}
}

func TestGT(t *testing.T) {
	ops := map[OperatorType]func(field string, value any, relationFields ...string) *Predicate{
		OpGT:  GT,
		OpGTE: GTE,
		OpLT:  LT,
		OpLTE: LTE,
	}

	for op, fn := range ops {
		tests := []struct {
			field string
			value any
			want  *Predicate
		}{
			{"field1", 1, &Predicate{Field: "field1", Operator: op, Value: 1}},
			{"field2", "abc", &Predicate{Field: "field2", Operator: op, Value: "abc"}},
			{"field3", true, &Predicate{Field: "field3", Operator: op, Value: true}},
		}

		for _, tt := range tests {
			if got := fn(tt.field, tt.value); !predicatesEqual(got, tt.want) {
				t.Errorf("%s(%q, %v) = %v, want %v", op.String(), tt.field, tt.value, got, tt.want)
			}
		}
	}
}

func TestLike(t *testing.T) {
	tests := []struct {
		field string
		value string
		want  *Predicate
	}{
		{"field1", "abc", &Predicate{Field: "field1", Operator: OpLIKE, Value: "abc"}},
		{"field2", "%def%", &Predicate{Field: "field2", Operator: OpLIKE, Value: "%def%"}},
		{"field3", "", &Predicate{Field: "field3", Operator: OpLIKE, Value: ""}},
	}

	for _, tt := range tests {
		if got := Like(tt.field, tt.value); !predicatesEqual(got, tt.want) {
			t.Errorf("Like(%q, %q) = %v, want %v", tt.field, tt.value, got, tt.want)
		}
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		field  string
		values []any
		want   *Predicate
	}{
		{"field1", []any{1, 2, 3}, &Predicate{Field: "field1", Operator: OpIN, Value: []any{1, 2, 3}}},
		{"field2", []any{"abc", "def"}, &Predicate{Field: "field2", Operator: OpIN, Value: []any{"abc", "def"}}},
		{"field3", []any{true, false}, &Predicate{Field: "field3", Operator: OpIN, Value: []any{true, false}}},
	}

	for _, tt := range tests {
		if got := In(tt.field, tt.values); !predicatesEqual(got, tt.want) {
			t.Errorf("In(%q, %v) = %v, want %v", tt.field, tt.values, got, tt.want)
		}
	}
}

func TestNotIn(t *testing.T) {
	tests := []struct {
		field  string
		values []any
		want   *Predicate
	}{
		{"field1", []any{1, 2, 3}, &Predicate{Field: "field1", Operator: OpNIN, Value: []any{1, 2, 3}}},
		{"field2", []any{"abc", "def"}, &Predicate{Field: "field2", Operator: OpNIN, Value: []any{"abc", "def"}}},
		{"field3", []any{true, false}, &Predicate{Field: "field3", Operator: OpNIN, Value: []any{true, false}}},
		{"field4", []any{}, &Predicate{Field: "field4", Operator: OpNIN, Value: []any{}}},
	}

	for _, tt := range tests {
		if got := NotIn(tt.field, tt.values); !predicatesEqual(got, tt.want) {
			t.Errorf("NotIn(%q, %v) = %v, want %v", tt.field, tt.values, got, tt.want)
		}
	}
}

func TestNull(t *testing.T) {
	tests := []struct {
		field string
		value bool
		want  *Predicate
	}{
		{"field1", true, &Predicate{Field: "field1", Operator: OpNULL, Value: true}},
		{"field2", false, &Predicate{Field: "field2", Operator: OpNULL, Value: false}},
	}

	for _, tt := range tests {
		if got := Null(tt.field, tt.value); !predicatesEqual(got, tt.want) {
			t.Errorf("Null(%q, %v) = %v, want %v", tt.field, tt.value, got, tt.want)
		}
	}
}

func TestIsFalse(t *testing.T) {
	values := map[bool]func(field string, relationFields ...string) *Predicate{
		true:  IsTrue,
		false: IsFalse,
	}

	for val, fn := range values {
		tests := []struct {
			field string
			want  *Predicate
		}{
			{"field1", &Predicate{Field: "field1", Operator: OpEQ, Value: val}},
			{"field2", &Predicate{Field: "field2", Operator: OpEQ, Value: val}},
			{"field3", &Predicate{Field: "field3", Operator: OpEQ, Value: val}},
		}

		for _, tt := range tests {
			if got := fn(tt.field); !predicatesEqual(got, tt.want) {
				t.Errorf("IsFalse(%q) = %v, want %v", tt.field, got, tt.want)
			}
		}
	}
}

func predicatesEqual(a, b *Predicate) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if a.Field != b.Field || a.Operator != b.Operator {
		return false
	}
	if !assert.ObjectsAreEqual(a.Value, b.Value) {
		return false
	}
	if len(a.And) != len(b.And) || len(a.Or) != len(b.Or) {
		return false
	}
	for i := range a.And {
		if !predicatesEqual(a.And[i], b.And[i]) {
			return false
		}
	}
	for i := range a.Or {
		if !predicatesEqual(a.Or[i], b.Or[i]) {
			return false
		}
	}
	return true
}
