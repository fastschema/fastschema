package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestNewSyncMap(t *testing.T) {
	sm := fs.NewSyncMap[int, string]()
	assert.NotNil(t, sm)
	assert.Equal(t, 0, sm.Len())
}

func TestSyncMapStore(t *testing.T) {
	sm := &fs.SyncMap[int, string]{}
	key := 42
	value := "test value"
	sm.Store(key, value)
	loadedValue, ok := sm.Load(key)
	assert.True(t, ok)
	assert.Equal(t, value, loadedValue)
}

func TestSyncMapLoad(t *testing.T) {
	sm := &fs.SyncMap[int, string]{}
	key := 42
	value := "test value"
	sm.Store(key, value)

	loadedValue, ok := sm.Load(key)
	assert.True(t, ok)
	assert.Equal(t, value, loadedValue)

	nonExistentKey := 99
	loadedValue, ok = sm.Load(nonExistentKey)
	assert.False(t, ok)
	assert.Equal(t, "", loadedValue)
}

func TestSyncMapLoadOrStore(t *testing.T) {
	sm := &fs.SyncMap[int, string]{}
	key := 42
	value := "test value"

	// Test loading non-existent key
	loadedValue, loaded := sm.LoadOrStore(key, value)
	assert.False(t, loaded)
	assert.Equal(t, value, loadedValue)

	// Test loading existing key
	loadedValue, loaded = sm.LoadOrStore(key, "new value")
	assert.True(t, loaded)
	assert.Equal(t, value, loadedValue)
}

func TestSyncMapDelete(t *testing.T) {
	sm := &fs.SyncMap[int, string]{}
	key := 42
	value := "test value"
	sm.Store(key, value)
	sm.Delete(key)
	_, ok := sm.Load(key)
	assert.False(t, ok)
	assert.Equal(t, 0, sm.Len())
}

func TestSyncMapLen(t *testing.T) {
	sm := &fs.SyncMap[int, string]{}
	assert.Equal(t, 0, sm.Len())

	// Add some key-value pairs
	sm.Store(1, "value1")
	sm.Store(2, "value2")
	sm.Store(3, "value3")

	assert.Equal(t, 3, sm.Len())

	// Delete a key-value pair
	sm.Delete(2)

	assert.Equal(t, 2, sm.Len())
}

func TestSyncMapKeys(t *testing.T) {
	sm := &fs.SyncMap[int, string]{}
	// Add some key-value pairs
	sm.Store(1, "value1")
	sm.Store(2, "value2")
	sm.Store(3, "value3")
	// Get the keys
	keys := sm.Keys()
	// Check the length of the keys slice
	assert.Equal(t, 3, len(keys))
	// Check the values of the keys
	assert.Contains(t, keys, 1)
	assert.Contains(t, keys, 2)
	assert.Contains(t, keys, 3)
}
