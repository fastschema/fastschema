package fs

import (
	"testing"
)

func TestDiskBase_Name(t *testing.T) {
	db := DiskBase{DiskName: "TestDisk"}
	expected := "TestDisk"
	if name := db.Name(); name != expected {
		t.Errorf("Expected DiskName to be %s, got %s", expected, name)
	}
}

func TestDiskBase_IsAllowedMime(t *testing.T) {
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
