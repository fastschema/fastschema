package utils_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fastschema/fastschema/pkg/utils"
)

type TestStruct struct {
	Name string
}

func TestGetDereferencedType(t *testing.T) {
	// Test case 1: Dereference a pointer to a struct
	ptr := &TestStruct{Name: "John"}
	result := utils.GetDereferencedType(ptr)
	expected := TestStruct{Name: "John"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 2: Dereference a pointer to a pointer to a struct
	ptrPtr := &ptr
	result = utils.GetDereferencedType(ptrPtr)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 3: Dereference a non-pointer value
	value := TestStruct{Name: "John"}
	result = utils.GetDereferencedType(value)
	if !reflect.DeepEqual(result, value) {
		t.Errorf("Expected %v, but got %v", value, result)
	}

	// Test case 4: Dereference a pointer to itself
	selfPtr := &ptrPtr
	result = utils.GetDereferencedType(selfPtr)
	if !reflect.DeepEqual(result, *ptr) {
		t.Errorf("Expected %v, but got %v", selfPtr, result)
	}

	// Test case 5: Dereference a pointer to a pointer to itself
	var a any = 5
	a = &a
	a = &a
	result = utils.GetDereferencedType(&a)
	if !reflect.DeepEqual(result, a) {
		t.Errorf("Expected %v, but got %v", **ptrPtr, result)
	}

	// Test case 6: Dereference a nil pointer
	var nilPtr *TestStruct
	result = utils.GetDereferencedType(nilPtr)
	if !reflect.DeepEqual(result, TestStruct{}) {
		t.Errorf("Expected %v, but got %v", TestStruct{}, result)
	}

	// Test case 7: Dereference a nil value
	var nilValue any
	result = utils.GetDereferencedType(nilValue)
	if !reflect.DeepEqual(result, nilValue) {
		t.Errorf("Expected %v, but got %v", nilValue, result)
	}
}

func TestGeneratePointerChain(t *testing.T) {
	result := utils.GeneratePointerChain(5, 5)
	expected := "*****int"
	if !strings.Contains(spew.Sdump(result), expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestDereferenceable(t *testing.T) {
	// Test case 1: Dereferenceable value
	value := "Hello"
	result := utils.Dereferenceable(value)
	expected := true
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 2: Non-dereferenceable value
	var ptr *int
	result = utils.Dereferenceable(ptr)
	expected = true
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 3: Nil value
	var nilValue any
	result = utils.Dereferenceable(nilValue)
	expected = false
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestIsNotAny(t *testing.T) {
	// Test case 1: Dereferenceable value that is not an interface
	value := "Hello"
	result := utils.IsNotAny(value)
	expected := true
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 2: Non-dereferenceable value
	var intf any
	result = utils.IsNotAny(intf)
	expected = false
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 3: Pointer value
	var ptr *int
	result = utils.IsNotAny(ptr)
	expected = true
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 4: Non-pointer value
	var nonPtr int
	result = utils.IsNotAny(nonPtr)
	expected = true
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 5: Nil value
	var nilValue any
	result = utils.IsNotAny(nilValue)
	expected = false
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 6: Pointer to any
	var any any
	result = utils.IsNotAny(&any)
	expected = false
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
