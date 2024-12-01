package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	// Test case 1
	slice1 := []int{1, 2, 3, 4, 5}
	mapper1 := func(x int) int {
		return x * 2
	}
	expected1 := []int{2, 4, 6, 8, 10}
	result1 := Map(slice1, mapper1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	slice2 := []string{"apple", "banana", "cherry"}
	mapper2 := func(s string) int {
		return len(s)
	}
	expected2 := []int{5, 6, 6}
	result2 := Map(slice2, mapper2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	slice3 := []float64{1.5, 2.5, 3.5}
	mapper3 := func(f float64) string {
		return fmt.Sprintf("%.2f", f)
	}
	expected3 := []string{"1.50", "2.50", "3.50"}
	result3 := Map(slice3, mapper3)
	assert.Equal(t, expected3, result3)
}

func TestFilter(t *testing.T) {
	// Test case 1
	slice1 := []int{1, 2, 3, 4, 5}
	predicate1 := func(x int) bool {
		return x%2 == 0
	}
	expected1 := []int{2, 4}
	result1 := Filter(slice1, predicate1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	slice2 := []string{"apple", "banana", "cherry"}
	predicate2 := func(s string) bool {
		return len(s) > 5
	}
	expected2 := []string{"banana", "cherry"}
	result2 := Filter(slice2, predicate2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	slice3 := []float64{1.5, 2.5, 3.5}
	predicate3 := func(f float64) bool {
		return f > 2.0
	}
	expected3 := []float64{2.5, 3.5}
	result3 := Filter(slice3, predicate3)
	assert.Equal(t, expected3, result3)
}

func TestContains(t *testing.T) {
	// Test case 1
	slice1 := []int{1, 2, 3, 4, 5}
	element1 := 3
	expected1 := true
	result1 := Contains(slice1, element1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	slice2 := []string{"apple", "banana", "cherry"}
	element2 := "orange"
	expected2 := false
	result2 := Contains(slice2, element2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	slice3 := []float64{1.5, 2.5, 3.5}
	element3 := 2.5
	expected3 := true
	result3 := Contains(slice3, element3)
	assert.Equal(t, expected3, result3)
}

func TestSliceEqual(t *testing.T) {
	// Test case 1
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{1, 2, 3, 4, 5}
	expected1 := true
	result1 := SliceEqual(slice1, slice2)
	assert.Equal(t, expected1, result1)

	// Test case 2
	slice3 := []string{"apple", "banana", "cherry"}
	slice4 := []string{"apple", "banana", "cherry"}
	expected2 := true
	result2 := SliceEqual(slice3, slice4)
	assert.Equal(t, expected2, result2)

	// Test case 3
	slice5 := []float64{1.5, 2.5, 3.5}
	slice6 := []float64{1.5, 2.5, 3.5}
	expected3 := true
	result3 := SliceEqual(slice5, slice6)
	assert.Equal(t, expected3, result3)

	// Test case 4
	slice7 := []int{1, 2, 3, 4, 5}
	slice8 := []int{1, 2, 3, 4}
	expected4 := false
	result4 := SliceEqual(slice7, slice8)
	assert.Equal(t, expected4, result4)

	// Test case 5
	slice9 := []string{"apple", "banana", "cherry"}
	slice10 := []string{"apple", "banana"}
	expected5 := false
	result5 := SliceEqual(slice9, slice10)
	assert.Equal(t, expected5, result5)

	// Test case 6
	slice11 := []float64{1.5, 2.5, 3.5}
	slice12 := []float64{1.5, 2.5}
	expected6 := false
	result6 := SliceEqual(slice11, slice12)
	assert.Equal(t, expected6, result6)

	// Test not equal slices
	slice13 := []int{1, 2, 3, 4, 5}
	slice14 := []int{1, 2, 3, 4, 6}
	expected7 := false
	result7 := SliceEqual(slice13, slice14)
	assert.Equal(t, expected7, result7)
}

func TestSliceInsertBeforeElement(t *testing.T) {
	// Test case 1
	slice1 := []int{1, 2, 3, 4, 5}
	newElement1 := 10
	checkIndexFn1 := func(element int) bool {
		return element == 3
	}
	expected1 := []int{1, 2, 10, 3, 4, 5}
	result1 := SliceInsertBeforeElement(slice1, newElement1, checkIndexFn1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	slice2 := []string{"apple", "banana", "cherry"}
	newElement2 := "orange"
	checkIndexFn2 := func(element string) bool {
		return element == "banana"
	}
	expected2 := []string{"apple", "orange", "banana", "cherry"}
	result2 := SliceInsertBeforeElement(slice2, newElement2, checkIndexFn2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	slice3 := []float64{1.5, 2.5, 3.5}
	newElement3 := 2.0
	checkIndexFn3 := func(element float64) bool {
		return element > 2.0
	}
	expected3 := []float64{1.5, 2.5, 2.0, 3.5}
	result3 := SliceInsertBeforeElement(slice3, newElement3, checkIndexFn3)
	assert.Equal(t, expected3, result3)

	// Test case 4
	slice4 := []int{1, 2, 3, 4, 5}
	newElement4 := 10
	checkIndexFn4 := func(element int) bool {
		return element == 6
	}
	expected4 := []int{1, 2, 3, 4, 5, 10}
	result4 := SliceInsertBeforeElement(slice4, newElement4, checkIndexFn4)
	assert.Equal(t, expected4, result4)

	// Test case 5
	slice5 := []string{"apple", "banana", "cherry"}
	newElement5 := "orange"
	checkIndexFn5 := func(element string) bool {
		return element == "mango"
	}
	expected5 := []string{"apple", "banana", "cherry", "orange"}
	result5 := SliceInsertBeforeElement(slice5, newElement5, checkIndexFn5)
	assert.Equal(t, expected5, result5)

	// Test case 6
	slice6 := []float64{1.5, 2.5, 3.5}
	newElement6 := 2.0
	checkIndexFn6 := func(element float64) bool {
		return element < 1.0
	}
	expected6 := []float64{1.5, 2.5, 3.5, 2.0}
	result6 := SliceInsertBeforeElement(slice6, newElement6, checkIndexFn6)
	assert.Equal(t, expected6, result6)
}

func TestIf(t *testing.T) {
	// Test case 1
	result1 := If(true, 10, 20)
	assert.Equal(t, 10, result1)

	// Test case 2
	result2 := If(false, "true", "false")
	assert.Equal(t, "false", result2)

	// Test case 3
	result3 := If(5 > 3, 1.5, 2.5)
	assert.Equal(t, 1.5, result3)
}

func TestIfFn(t *testing.T) {
	// Test case 1
	result1 := IfFn(true, func() int { return 10 }, func() int { return 20 })
	assert.Equal(t, 10, result1)

	// Test case 2
	result2 := IfFn(false, func() string { return "true" }, func() string { return "false" })
	assert.Equal(t, "false", result2)

	// Test case 3
	result3 := IfFn(5 > 3, func() float64 { return 1.5 }, func() float64 { return 2.5 })
	assert.Equal(t, 1.5, result3)
}

func TestGetMapKeys(t *testing.T) {
	// Test case 1
	map1 := map[int]string{1: "apple", 2: "banana", 3: "cherry"}
	expected1 := []int{1, 2, 3}
	result1 := MapKeys(map1)
	assert.ElementsMatch(t, expected1, result1)

	// Test case 2
	map2 := map[string]int{"apple": 1, "banana": 2, "cherry": 3}
	expected2 := []string{"apple", "banana", "cherry"}
	result2 := MapKeys(map2)
	assert.ElementsMatch(t, expected2, result2)

	// Test case 3
	map3 := map[float64]bool{1.5: true, 2.5: false, 3.5: true}
	expected3 := []float64{1.5, 2.5, 3.5}
	result3 := MapKeys(map3)
	assert.ElementsMatch(t, expected3, result3)
}

func TestGetMapValues(t *testing.T) {
	// Test case 1
	map1 := map[int]string{1: "apple", 2: "banana", 3: "cherry"}
	expected1 := []string{"apple", "banana", "cherry"}
	result1 := MapValues(map1)
	assert.ElementsMatch(t, expected1, result1)

	// Test case 2
	map2 := map[string]int{"apple": 1, "banana": 2, "cherry": 3}
	expected2 := []int{1, 2, 3}
	result2 := MapValues(map2)
	assert.ElementsMatch(t, expected2, result2)

	// Test case 3
	map3 := map[float64]bool{1.5: true, 2.5: false, 3.5: true}
	expected3 := []bool{true, false, true}
	result3 := MapValues(map3)
	assert.ElementsMatch(t, expected3, result3)
}

func TestPick(t *testing.T) {
	// Test case 1: Valid nested map path
	obj1 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path1 := "foo.bar"
	expected1 := []int{1, 2, 3}
	result1 := Pick(obj1, path1, 0)
	assert.Equal(t, expected1, result1)

	// Test case 2: Invalid path, default value provided
	obj2 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path2 := "foo.baz"
	expected2 := 0
	result2 := Pick(obj2, path2, 0)
	assert.Equal(t, expected2, result2)

	// Test case 3: Invalid array index, default value provided
	obj3 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path3 := "foo.bar.3"
	expected3 := 0
	result3 := Pick(obj3, path3, 0)
	assert.Equal(t, expected3, result3)

	// Test case 4: Valid nested map path, no default value provided
	obj4 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path4 := "foo.bar"
	expected4 := []int{1, 2, 3}
	result4 := Pick(obj4, path4)
	assert.Equal(t, expected4, result4)

	// Test case 5: Invalid path, no default value provided
	obj5 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path5 := "foo.baz.qux"
	var expected5 any = nil
	result5 := Pick(obj5, path5)
	assert.Equal(t, expected5, result5)

	// Test case 6: Valid array index
	obj6 := map[string]any{
		"foo": []any{map[string]any{"bar": "baz"}},
	}
	path6 := "foo.0.bar"
	expected6 := "baz"
	result6 := Pick(obj6, path6, "default")
	assert.Equal(t, expected6, result6)

	// Test case 7: Invalid array index, default value provided
	obj7 := map[string]any{
		"foo": []any{map[string]any{"bar": "baz"}},
	}
	path7 := "foo.1.bar"
	expected7 := "default"
	result7 := Pick(obj7, path7, "default")
	assert.Equal(t, expected7, result7)
}

func TestEscapeQuery(t *testing.T) {
	// Test case 1
	query1 := "SELECT * FROM users WHERE name = 'John Doe'"
	expected1 := `SELECT \* FROM users WHERE name = 'John Doe'$`
	result1 := EscapeQuery(query1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	query2 := "INSERT INTO products (name, price) VALUES ('Apple', 1.99)"
	expected2 := `INSERT INTO products \(name, price\) VALUES \('Apple', 1\.99\)$`
	result2 := EscapeQuery(query2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	query3 := "UPDATE users SET name = 'Jane Doe' WHERE id = 1"
	expected3 := `UPDATE users SET name = 'Jane Doe' WHERE id = 1$`
	result3 := EscapeQuery(query3)
	assert.Equal(t, expected3, result3)

	// Test case 4
	query4 := "DELETE FROM products WHERE price > 10.0"
	expected4 := `DELETE FROM products WHERE price > 10\.0$`
	result4 := EscapeQuery(query4)
	assert.Equal(t, expected4, result4)
}

func TestIsNumber(t *testing.T) {
	// Test case 1: int
	result1 := IsNumber(10)
	assert.True(t, result1)

	// Test case 2: float64
	result2 := IsNumber(3.14)
	assert.True(t, result2)

	// Test case 3: string
	result3 := IsNumber("123")
	assert.False(t, result3)

	// Test case 4: bool
	result4 := IsNumber(true)
	assert.False(t, result4)

	// Test case 5: struct
	type Person struct {
		Name string
		Age  int
	}
	person := Person{Name: "John", Age: 30}
	result5 := IsNumber(person)
	assert.False(t, result5)
}

func TestMust(t *testing.T) {
	// Test case 1: value is not nil
	value1 := 10
	var err1 error = nil
	result1 := Must(value1, err1)
	assert.Equal(t, value1, result1)

	// Test case 2: value is nil
	var value2 any = nil
	err2 := errors.New("some error")
	assert.Panics(t, func() {
		Must(value2, err2)
	})
}

func TestIsValidBool(t *testing.T) {
	// Test case 1: Valid bool value
	result1 := IsValidBool(true)
	assert.True(t, result1)

	// Test case 2: Invalid bool value
	result2 := IsValidBool("true")
	assert.False(t, result2)

	// Test case 3: Invalid bool value
	result3 := IsValidBool(10)
	assert.False(t, result3)

	// Test case 4: Invalid bool value
	result4 := IsValidBool(3.14)
	assert.False(t, result4)
}

func TestIsValidTime(t *testing.T) {
	// Test case 1: Valid RFC3339 time
	validTime1 := "2022-01-01T12:00:00Z"
	result1 := IsValidTime(validTime1)
	assert.True(t, result1)

	// Test case 2: Valid custom time format
	validTime2 := "2022-01-01 12:00:00"
	result2 := IsValidTime(validTime2)
	assert.True(t, result2)

	// Test case 3: Invalid time format
	invalidTime := "2022-01-01T12:00:00"
	result3 := IsValidTime(invalidTime)
	assert.False(t, result3)

	// Test case 4: Invalid type
	invalidType := 123
	result4 := IsValidTime(invalidType)
	assert.False(t, result4)
}

func TestIsValidString(t *testing.T) {
	// Test case 1: Valid string
	result1 := IsValidString("hello")
	assert.True(t, result1)

	// Test case 2: Invalid string
	result2 := IsValidString(123)
	assert.False(t, result2)

	// Test case 3: Empty string
	result3 := IsValidString("")
	assert.True(t, result3)

	// Test case 4: Nil value
	result4 := IsValidString(nil)
	assert.False(t, result4)

	// Test case 5: Non-string type
	result5 := IsValidString(3.14)
	assert.False(t, result5)
}

func TestIsValidFloat(t *testing.T) {
	// Test case 1: float32
	result1 := IsValidFloat(float32(3.14))
	assert.True(t, result1)

	// Test case 2: float64
	result2 := IsValidFloat(float64(3.14))
	assert.True(t, result2)

	// Test case 3: valid string representation of float
	result3 := IsValidFloat("3.14")
	assert.True(t, result3)

	// Test case 4: invalid string representation of float
	result4 := IsValidFloat("abc")
	assert.False(t, result4)

	// Test case 5: valid type int
	result5 := IsValidFloat(10)
	assert.True(t, result5)
}

func TestIsValidInt(t *testing.T) {
	// Test case 1: int
	result1 := IsValidInt(10)
	assert.True(t, result1)

	// Test case 2: int8
	result2 := IsValidInt(int8(10))
	assert.True(t, result2)

	// Test case 3: int16
	result3 := IsValidInt(int16(10))
	assert.True(t, result3)

	// Test case 4: int32
	result4 := IsValidInt(int32(10))
	assert.True(t, result4)

	// Test case 5: int64
	result5 := IsValidInt(int64(10))
	assert.True(t, result5)

	// Test case 6: uint
	result6 := IsValidInt(uint(10))
	assert.True(t, result6)

	// Test case 7: uint8
	result7 := IsValidInt(uint8(10))
	assert.True(t, result7)

	// Test case 8: uint16
	result8 := IsValidInt(uint16(10))
	assert.True(t, result8)

	// Test case 9: uint32
	result9 := IsValidInt(uint32(10))
	assert.True(t, result9)

	// Test case 10: uint64
	result10 := IsValidInt(uint64(10))
	assert.True(t, result10)

	// Test case 11: float64
	result11 := IsValidInt(3.14)
	assert.False(t, result11)

	// Test case 12: string
	result12 := IsValidInt("10")
	assert.True(t, result12)

	// Test case 13: bool
	result13 := IsValidInt(true)
	assert.False(t, result13)
}

func TestIsValidUInt(t *testing.T) {
	// Test case 1: uint
	result1 := IsValidUInt(uint(10))
	assert.True(t, result1)

	// Test case 2: uint8
	result2 := IsValidUInt(uint8(10))
	assert.True(t, result2)

	// Test case 3: uint16
	result3 := IsValidUInt(uint16(10))
	assert.True(t, result3)

	// Test case 4: uint32
	result4 := IsValidUInt(uint32(10))
	assert.True(t, result4)

	// Test case 5: uint64
	result5 := IsValidUInt(uint64(10))
	assert.True(t, result5)

	// Test case 6: int (positive)
	result6 := IsValidUInt(int(10))
	assert.True(t, result6)

	// Test case 7: int (negative)
	result7 := IsValidUInt(int(-10))
	assert.False(t, result7)

	// Test case 8: int8 (positive)
	result8 := IsValidUInt(int8(10))
	assert.True(t, result8)

	// Test case 9: int8 (negative)
	result9 := IsValidUInt(int8(-10))
	assert.False(t, result9)

	// Test case 10: int16 (positive)
	result10 := IsValidUInt(int16(10))
	assert.True(t, result10)

	// Test case 11: int16 (negative)
	result11 := IsValidUInt(int16(-10))
	assert.False(t, result11)

	// Test case 12: int32 (positive)
	result12 := IsValidUInt(int32(10))
	assert.True(t, result12)

	// Test case 13: int32 (negative)
	result13 := IsValidUInt(int32(-10))
	assert.False(t, result13)

	// Test case 14: int64 (positive)
	result14 := IsValidUInt(int64(10))
	assert.True(t, result14)

	// Test case 15: int64 (negative)
	result15 := IsValidUInt(int64(-10))
	assert.False(t, result15)

	// Test case 16: float32 (positive integer)
	result16 := IsValidUInt(float32(10.0))
	assert.True(t, result16)

	// Test case 17: float32 (positive non-integer)
	result17 := IsValidUInt(float32(10.5))
	assert.False(t, result17)

	// Test case 18: float32 (negative)
	result18 := IsValidUInt(float32(-10.0))
	assert.False(t, result18)

	// Test case 19: float64 (positive integer)
	result19 := IsValidUInt(float64(10.0))
	assert.True(t, result19)

	// Test case 20: float64 (positive non-integer)
	result20 := IsValidUInt(float64(10.5))
	assert.False(t, result20)

	// Test case 21: float64 (negative)
	result21 := IsValidUInt(float64(-10.0))
	assert.False(t, result21)

	// Test case 22: valid string representation of uint
	result22 := IsValidUInt("10")
	assert.True(t, result22)

	// Test case 23: invalid string representation of uint
	result23 := IsValidUInt("abc")
	assert.False(t, result23)

	// Test case 24: bool
	result24 := IsValidUInt(true)
	assert.False(t, result24)

	// Test case 25: struct
	type Person struct {
		Name string
		Age  int
	}
	person := Person{Name: "John", Age: 30}
	result25 := IsValidUInt(person)
	assert.False(t, result25)
}

func TestWriteFile(t *testing.T) {
	// Test case 1: Successful write
	filePath1 := t.TempDir() + "testfile.txt"
	content1 := "Hello, World!"
	err1 := WriteFile(filePath1, content1)
	assert.NoError(t, err1)

	// Verify that the file exists and contains the correct content
	file1, err2 := os.Open(filePath1)
	assert.NoError(t, err2)
	defer file1.Close()

	stat1, err3 := file1.Stat()
	assert.NoError(t, err3)
	assert.Equal(t, int64(len(content1)), stat1.Size())

	buf1 := make([]byte, len(content1))
	_, err4 := file1.Read(buf1)
	assert.NoError(t, err4)
	assert.Equal(t, content1, string(buf1))

	// Test case 2: Error creating file
	filePath2 := "/invalid/path/testfile.txt"
	content2 := "Hello, World!"
	err5 := WriteFile(filePath2, content2)
	assert.Error(t, err5)
	assert.True(t, errors.Is(err5, os.ErrNotExist))
}

func TestAppendFile(t *testing.T) {
	// Test case 1: Successful append
	filePath1 := t.TempDir() + "testfile.txt"
	initialContent1 := "Hello, World!"
	appendContent1 := " Goodbye, World!"
	expectedContent1 := initialContent1 + appendContent1

	// Write initial content to the file
	err := WriteFile(filePath1, initialContent1)
	assert.NoError(t, err)

	// Append content to the file
	err = AppendFile(filePath1, appendContent1)
	assert.NoError(t, err)

	// Verify that the file contains the correct content
	file1, err := os.Open(filePath1)
	assert.NoError(t, err)
	defer file1.Close()

	stat1, err := file1.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expectedContent1)), stat1.Size())

	buf1 := make([]byte, len(expectedContent1))
	_, err = file1.Read(buf1)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent1, string(buf1))

	// Test case 2: File does not exist, create and append
	filePath2 := t.TempDir() + "newfile.txt"
	appendContent2 := "Hello, New World!"
	expectedContent2 := appendContent2

	// Append content to the new file
	err = AppendFile(filePath2, appendContent2)
	assert.NoError(t, err)

	// Verify that the file contains the correct content
	file2, err := os.Open(filePath2)
	assert.NoError(t, err)
	defer file2.Close()

	stat2, err := file2.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expectedContent2)), stat2.Size())

	buf2 := make([]byte, len(expectedContent2))
	_, err = file2.Read(buf2)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent2, string(buf2))

	// Test case 3: Error opening file
	filePath3 := "/invalid/path/testfile.txt"
	appendContent3 := "Hello, World!"
	err = AppendFile(filePath3, appendContent3)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestIsFileExists(t *testing.T) {
	// Test case 1: Existing file
	filePath1 := t.TempDir() + "testfile.txt"
	err := WriteFile(filePath1, "Hello, World!")
	assert.NoError(t, err)
	result1 := IsFileExists(filePath1)
	assert.True(t, result1)

	// Test case 2: Non-existing file
	nonExistingFilePath := "/path/to/non-existing/file.txt"
	result2 := IsFileExists(nonExistingFilePath)
	assert.False(t, result2)
}

func TestCopyFile(t *testing.T) {
	src := t.TempDir() + "testfile.txt"
	dst := t.TempDir() + "testfile2.txt"
	err := WriteFile(src, "Hello, World!")
	assert.NoError(t, err)

	// Test case 1: Successful file copy
	err = CopyFile(src, dst)
	assert.NoError(t, err)

	// Test case 2: Source file does not exist
	err = CopyFile("nonexistent/source/file.txt", dst)
	assert.Error(t, err)

	// Test case 3: Destination file already exists
	err = CopyFile(src, "existing/destination/file.txt")
	assert.Error(t, err)
}

func TestMkDirs(t *testing.T) {
	// Test case 1: Create multiple directories successfully
	dir1 := t.TempDir() + "/dir1"
	dir2 := t.TempDir() + "/dir2"
	err := MkDirs(dir1, dir2)
	assert.NoError(t, err)

	// Verify that the directories exist
	_, err = os.Stat(dir1)
	assert.NoError(t, err)
	assert.True(t, os.IsNotExist(err) == false)

	_, err = os.Stat(dir2)
	assert.NoError(t, err)
	assert.True(t, os.IsNotExist(err) == false)

	// Test case 2: Create nested directories successfully
	nestedDir := t.TempDir() + "/parent/child"
	err = MkDirs(nestedDir)
	assert.NoError(t, err)

	// Verify that the nested directories exist
	_, err = os.Stat(nestedDir)
	assert.NoError(t, err)
	assert.True(t, os.IsNotExist(err) == false)

	// Test case 3: Error creating directory (invalid path)
	invalidDir := "/invalid/path/dir"
	err = MkDirs(invalidDir)
	assert.Error(t, err)
}

func TestEnv(t *testing.T) {
	// Test case 1: Environment variable exists
	os.Setenv("TEST_ENV", "test_value")
	expected1 := "test_value"
	result1 := Env("TEST_ENV")
	assert.Equal(t, expected1, result1)

	// Test case 2: Environment variable does not exist, default value provided
	expected2 := "default_value"
	result2 := Env("NON_EXISTING_ENV", "default_value")
	assert.Equal(t, expected2, result2)

	// Test case 3: Environment variable does not exist, no default value provided
	expected3 := ""
	result3 := Env("NON_EXISTING_ENV")
	assert.Equal(t, expected3, result3)
}

func TestReadCloserToString(t *testing.T) {
	// Test case 1: ReadCloser with valid data
	rc1 := io.NopCloser(strings.NewReader("Hello, World!"))
	expected1 := "Hello, World!"
	result1, err1 := ReadCloserToString(rc1)
	assert.NoError(t, err1)
	assert.Equal(t, expected1, result1)

	// Test case 2: ReadCloser with empty data
	rc2 := io.NopCloser(strings.NewReader(""))
	expected2 := ""
	result2, err2 := ReadCloserToString(rc2)
	assert.NoError(t, err2)
	assert.Equal(t, expected2, result2)

	// Test case 3: ReadCloser with error
	rc3 := io.NopCloser(errorReader{})
	expected3 := ""
	result3, err3 := ReadCloserToString(rc3)
	assert.Error(t, err3)
	assert.Equal(t, expected3, result3)
}

// Helper struct for generating ReadCloser with error
type errorReader struct{}

func (er errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error reading data")
}

func (er errorReader) Close() error {
	return nil
}

func TestCapitalize(t *testing.T) {
	// Test case 1
	input1 := "hello"
	expected1 := "Hello"
	result1 := Capitalize(input1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	input2 := "world"
	expected2 := "World"
	result2 := Capitalize(input2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	input3 := "foo bar"
	expected3 := "Foo bar"
	result3 := Capitalize(input3)
	assert.Equal(t, expected3, result3)

	// Test case 4
	input4 := ""
	expected4 := ""
	result4 := Capitalize(input4)
	assert.Equal(t, expected4, result4)
}

func TestCreateSwaggerUIPage(t *testing.T) {
	specURL := "https://example.com/swagger.json"
	result := CreateSwaggerUIPage(specURL)
	assert.Contains(t, result, specURL)
}

func TestTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"hello_world", "Hello World"},
		{"hello-world", "Hello World"},
		{"hello   world", "Hello World"},
		{"", ""},
	}

	for _, test := range tests {
		result := Title(test.input)
		if result != test.expected {
			t.Errorf("Title(%s) = %s; expected %s", test.input, result, test.expected)
		}
	}
}

func TestCreateSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "HelloWorld",
			expected: "hello_world",
		},
		{
			input:    "FastSchema",
			expected: "fast_schema",
		},
		{
			input:    "CreateSnakeCase",
			expected: "create_snake_case",
		},
		{
			input:    "APIEndpoint",
			expected: "api_endpoint",
		},
		{
			input:    "UserID",
			expected: "user_id",
		},
		{
			input:    "HTTPStatusCode",
			expected: "http_status_code",
		},
	}

	for _, test := range tests {
		result := ToSnakeCase(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestGetStructFieldName(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field_1"`
		Field2 int    `json:"field_2,omitempty"`
		Field3 bool   `json:"-"`
	}

	field1Name := GetStructFieldName(reflect.TypeOf(TestStruct{}).Field(0))
	assert.Equal(t, "field_1", field1Name)

	field2Name := GetStructFieldName(reflect.TypeOf(TestStruct{}).Field(1))
	assert.Equal(t, "field_2", field2Name)

	field3Name := GetStructFieldName(reflect.TypeOf(TestStruct{}).Field(2))
	assert.Equal(t, "", field3Name)
}

func TestParseStructFieldTag(t *testing.T) {
	type TestStruct struct {
		Field1 string `fs:"name=aaaaa"`
		Field2 int    `fs:"type=text;name=title;multiple;size=10;unique;;sortable;"`
	}

	expected1 := map[string]string{"name": "aaaaa"}
	expected2 := map[string]string{
		"type":     "text",
		"name":     "title",
		"multiple": "",
		"size":     "10",
		"unique":   "",
		"sortable": "",
	}

	result1 := ParseStructFieldTag(reflect.TypeOf(TestStruct{}).Field(0), "fs")
	result2 := ParseStructFieldTag(reflect.TypeOf(TestStruct{}).Field(1), "fs")

	assert.Equal(t, expected1, result1)
	assert.Equal(t, expected2, result2)
}

func TestParseHJSON(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	// Case 1: Invalid input
	_, err := ParseHJSON[testStruct]([]byte(`{`))
	assert.Error(t, err)

	// Case 2: Valid input
	result, err := ParseHJSON[testStruct]([]byte(`{
		"name": "John",
		"age": 30,
		"email": "john@example.com"
	}`))

	assert.NoError(t, err)
	assert.Equal(t, testStruct{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}, result)
}

func TestSendRequest(t *testing.T) {
	type TR struct {
		Message string `json:"message"`
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	// Case 1: Missing protocol scheme
	_, err := SendRequest[TR]("GET", "://example.local", headers, nil)
	assert.ErrorContains(t, err, "missing protocol scheme")

	// Case 2: Timeout
	backUpClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Millisecond)
			},
		},
	}
	_, err = SendRequest[TR]("GET", "http://example.local", headers, nil)
	assert.ErrorContains(t, err, "timeout")
	http.DefaultClient = backUpClient

	// Case 3: Access token server error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()
	_, err = SendRequest[TR]("GET", errorServer.URL, headers, nil)
	assert.ErrorContains(t, err, "request failed with status code")

	// Case 4: Invalid JSON response
	errorServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer errorServer.Close()
	_, err = SendRequest[TR]("GET", errorServer.URL, headers, nil)
	assert.ErrorContains(t, err, "invalid character")

	// Case 5: Successful request
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	}))
	defer successServer.Close()
	resp, err := SendRequest[TR]("GET", successServer.URL, headers, nil)
	assert.NoError(t, err)
	assert.Equal(t, TR{Message: "success"}, resp)
}

func TestIsValidEmail(t *testing.T) {
	// Test case 1: Valid email
	result1 := IsValidEmail("test@example.com")
	assert.True(t, result1)

	// Test case 2: Invalid email (missing @)
	result2 := IsValidEmail("testexample.com")
	assert.False(t, result2)

	// Test case 3: Invalid email (missing domain)
	result3 := IsValidEmail("test@.com")
	assert.False(t, result3)

	// Test case 4: Invalid email (missing top-level domain)
	result4 := IsValidEmail("test@example")
	assert.False(t, result4)

	// Test case 5: Invalid email (special characters)
	result5 := IsValidEmail("test@exa!mple.com")
	assert.False(t, result5)

	// Test case 6: Invalid email (spaces)
	result6 := IsValidEmail("test@ example.com")
	assert.False(t, result6)

	// Test case 7: Invalid email (integer input)
	result7 := IsValidEmail(12345)
	assert.False(t, result7)

	// Test case 8: Invalid email (nil input)
	result8 := IsValidEmail(nil)
	assert.False(t, result8)

	// Test case 9: Valid email with subdomain
	result9 := IsValidEmail("test@mail.example.com")
	assert.True(t, result9)

	// Test case 10: Valid email with plus sign
	result10 := IsValidEmail("test+label@example.com")
	assert.True(t, result10)
}
