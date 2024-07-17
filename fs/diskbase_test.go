package fs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskBaseName(t *testing.T) {
	db := DiskBase{DiskName: "TestDisk"}
	expected := "TestDisk"
	if name := db.Name(); name != expected {
		t.Errorf("Expected DiskName to be %s, got %s", expected, name)
	}
}

func TestDiskBaseIsAllowedMime(t *testing.T) {
	db := DiskBase{}

	tests := []struct {
		mime     string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"text/plain", true},
		{"application/json", false}, // Not in the allowed list
	}

	for _, test := range tests {
		if allowed := db.IsAllowedMime(test.mime); allowed != test.expected {
			t.Errorf("Expected %t for mime type %s, got %t", test.expected, test.mime, allowed)
		}
	}
}

func TestBaseRcloneDiskName(t *testing.T) {
	r := &DiskBase{
		DiskName: "mydisk",
	}

	name := r.Name()
	assert.Equal(t, "mydisk", name)
}

func TestDiskBaseUploadFilePath(t *testing.T) {
	db := DiskBase{
		DiskName: "TestDisk",
		Root:     "/path/to/root",
	}

	tests := []string{
		"file.jpg",
		"document.pdf",
		"script.sh",
	}

	for _, test := range tests {
		path := db.UploadFilePath(test)
		assert.True(t, strings.HasSuffix(path, "_"+test))
		assert.True(t, strings.HasPrefix(path, "/path/to/root"))
	}
}
