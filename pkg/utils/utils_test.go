package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	length := 10
	randomString := RandomString(length)
	assert.Equal(t, length, len(randomString))
}

func TestSecureRandomBytes(t *testing.T) {
	length := 10
	randomBytes, err := SecureRandomBytes(length)
	assert.NoError(t, err)
	assert.Equal(t, length, len(randomBytes))
}

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
	result1 := GetMapKeys(map1)
	assert.ElementsMatch(t, expected1, result1)

	// Test case 2
	map2 := map[string]int{"apple": 1, "banana": 2, "cherry": 3}
	expected2 := []string{"apple", "banana", "cherry"}
	result2 := GetMapKeys(map2)
	assert.ElementsMatch(t, expected2, result2)

	// Test case 3
	map3 := map[float64]bool{1.5: true, 2.5: false, 3.5: true}
	expected3 := []float64{1.5, 2.5, 3.5}
	result3 := GetMapKeys(map3)
	assert.ElementsMatch(t, expected3, result3)
}

func TestGetMapValues(t *testing.T) {
	// Test case 1
	map1 := map[int]string{1: "apple", 2: "banana", 3: "cherry"}
	expected1 := []string{"apple", "banana", "cherry"}
	result1 := GetMapValues(map1)
	assert.ElementsMatch(t, expected1, result1)

	// Test case 2
	map2 := map[string]int{"apple": 1, "banana": 2, "cherry": 3}
	expected2 := []int{1, 2, 3}
	result2 := GetMapValues(map2)
	assert.ElementsMatch(t, expected2, result2)

	// Test case 3
	map3 := map[float64]bool{1.5: true, 2.5: false, 3.5: true}
	expected3 := []bool{true, false, true}
	result3 := GetMapValues(map3)
	assert.ElementsMatch(t, expected3, result3)
}

func TestPick(t *testing.T) {
	// Test case 1
	obj1 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path1 := "foo.bar"
	expected1 := []int{1, 2, 3}
	result1 := Pick(obj1, path1, 0)
	assert.Equal(t, expected1, result1)

	// Test case 2
	obj2 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path2 := "foo.baz"
	var expected2 any = nil
	result2 := Pick(obj2, path2, nil)
	assert.Equal(t, expected2, result2)

	// Test case 3
	obj3 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path3 := "foo.bar.3"
	var expected3 any = nil
	result3 := Pick(obj3, path3, nil)
	assert.Equal(t, expected3, result3)

	// Test case 4
	obj4 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path4 := "foo.bar"
	expected4 := []int{1, 2, 3}
	result4 := Pick(obj4, path4, nil)
	assert.Equal(t, expected4, result4)

	// Test case 5
	obj5 := map[string]any{
		"foo": map[string]any{
			"bar": []int{1, 2, 3},
		},
	}
	path5 := "foo.baz.qux"
	var expected5 any = nil
	result5 := Pick(obj5, path5, nil)
	assert.Equal(t, expected5, result5)
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
	// Test case 1: Valid uint
	result1 := IsValidUInt(uint(10))
	assert.True(t, result1)

	// Test case 2: Valid uint8
	result2 := IsValidUInt(uint8(5))
	assert.True(t, result2)

	// Test case 3: Valid uint16
	result3 := IsValidUInt(uint16(100))
	assert.True(t, result3)

	// Test case 4: Valid uint32
	result4 := IsValidUInt(uint32(1000))
	assert.True(t, result4)

	// Test case 5: Valid uint64
	result5 := IsValidUInt(uint64(10000))
	assert.True(t, result5)

	// Test case 6: Valid int (positive)
	result6 := IsValidUInt(int(10))
	assert.True(t, result6)

	// Test case 7: Valid int8 (positive)
	result7 := IsValidUInt(int8(5))
	assert.True(t, result7)

	// Test case 8: Valid int16 (positive)
	result8 := IsValidUInt(int16(100))
	assert.True(t, result8)

	// Test case 9: Valid int32 (positive)
	result9 := IsValidUInt(int32(1000))
	assert.True(t, result9)

	// Test case 10: Valid int64 (positive)
	result10 := IsValidUInt(int64(10000))
	assert.True(t, result10)

	// Test case 11: Valid int (negative)
	result11 := IsValidUInt(int(-10))
	assert.False(t, result11)

	// Test case 12: Valid int8 (negative)
	result12 := IsValidUInt(int8(-5))
	assert.False(t, result12)

	// Test case 13: Valid int16 (negative)
	result13 := IsValidUInt(int16(-100))
	assert.False(t, result13)

	// Test case 14: Valid int32 (negative)
	result14 := IsValidUInt(int32(-1000))
	assert.False(t, result14)

	// Test case 15: Valid int64 (negative)
	result15 := IsValidUInt(int64(-10000))
	assert.False(t, result15)

	// Test case 16: Valid float (positive) but not an integer
	result16 := IsValidUInt(float64(10.5))
	assert.False(t, result16)

	// Test case 17: Valid float (negative)
	result17 := IsValidUInt(float64(-10.5))
	assert.False(t, result17)

	// Test case 18: Valid uint string
	result18 := IsValidUInt("10")
	assert.True(t, result18)

	// Test case 19: Invalid bool
	result19 := IsValidUInt(true)
	assert.False(t, result19)

	// Test case 20: Invalid slice
	result20 := IsValidUInt([]int{1, 2, 3})
	assert.False(t, result20)

	// Test case 21: Invalid map
	result21 := IsValidUInt(map[string]int{"a": 1, "b": 2})
	assert.False(t, result21)

	// Test case 22: Invalid struct
	type Person struct {
		Name string
		Age  int
	}
	result22 := IsValidUInt(Person{Name: "John", Age: 30})
	assert.False(t, result22)
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
	// Test case 1
	filePath1 := t.TempDir() + "testfile.txt"
	content1 := "Hello, World!"
	err1 := AppendFile(filePath1, content1)
	assert.NoError(t, err1)

	// Verify the content of the file
	file1, err2 := os.Open(filePath1)
	assert.NoError(t, err2)
	defer file1.Close()

	stat1, err3 := file1.Stat()
	assert.NoError(t, err3)

	fileSize1 := stat1.Size()
	buffer1 := make([]byte, fileSize1)
	_, err4 := file1.Read(buffer1)
	assert.NoError(t, err4)

	fileContent1 := string(buffer1)
	assert.Equal(t, content1, fileContent1)

	// Test case 2
	content2 := "Append more content"
	err5 := AppendFile(filePath1, content2)
	assert.NoError(t, err5)

	// Verify the content of the file
	file2, err6 := os.Open(filePath1)
	assert.NoError(t, err6)
	defer file2.Close()

	stat2, err7 := file2.Stat()
	assert.NoError(t, err7)

	fileSize2 := stat2.Size()
	buffer2 := make([]byte, fileSize2)
	_, err8 := file2.Read(buffer2)
	assert.NoError(t, err8)

	fileContent2 := string(buffer2)
	expectedContent2 := content1 + content2
	assert.Equal(t, expectedContent2, fileContent2)
}

func TestIsFileExists(t *testing.T) {
	// Test case 1: Existing file
	filePath1 := t.TempDir() + "testfile.txt"
	WriteFile(filePath1, "Hello, World!")
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
	WriteFile(src, "Hello, World!")

	// Test case 1: Successful file copy
	err := CopyFile(src, dst)
	assert.NoError(t, err)

	// Test case 2: Source file does not exist
	err = CopyFile("nonexistent/source/file.txt", dst)
	assert.Error(t, err)

	// Test case 3: Destination file already exists
	err = CopyFile(src, "existing/destination/file.txt")
	assert.Error(t, err)
}

func TestGenerateHash(t *testing.T) {
	// Test case 1: Non-empty input
	input1 := "password123"
	hash1, err1 := GenerateHash(input1)
	if err1 != nil {
		t.Errorf("Unexpected error: %v", err1)
	}

	assert.NotEmpty(t, hash1)
	// Test case 2: Empty input
	input2 := ""
	expectedErr2 := errors.New("hash: input cannot be empty")
	_, err2 := GenerateHash(input2)
	if err2 == nil {
		t.Error("Expected error, but got nil")
	}
	if err2.Error() != expectedErr2.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedErr2, err2)
	}
}

func TestCheckHash(t *testing.T) {
	input := "password"
	hash, err := GenerateHash(input)
	assert.NoError(t, err)
	err = CheckHash(input, hash)
	assert.NoError(t, err)

	// Test case with invalid hash
	invalidHash := "invalid_hash"
	err = CheckHash(input, invalidHash)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid hash")

	// Test case with incorrect input
	incorrectInput := "incorrect_password"
	err = CheckHash(incorrectInput, hash)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid hash")
}

func TestMkDirs(t *testing.T) {
	// Test case 1
	dir1 := t.TempDir() + "/path/to/dir1"
	err1 := MkDirs(dir1)
	assert.NoError(t, err1)
	_, err := os.Stat(dir1)
	assert.False(t, os.IsNotExist(err))

	// Test case 2
	dir2 := t.TempDir() + "/path/to/dir2"
	subdir2 := t.TempDir() + "/path/to/dir2/subdir"
	err2 := MkDirs(dir2, subdir2)
	assert.NoError(t, err2)
	_, err = os.Stat(dir2)
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(subdir2)
	assert.False(t, os.IsNotExist(err))

	// Test case 3
	dir3 := t.TempDir() + "/path/to/dir3"
	err3 := MkDirs(dir3)
	assert.NoError(t, err3)
	_, err = os.Stat(dir3)
	assert.False(t, os.IsNotExist(err))
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
