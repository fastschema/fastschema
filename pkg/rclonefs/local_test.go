package rclonefs

import (
	"os"
	"testing"

	"github.com/rclone/rclone/backend/local"
	"github.com/stretchr/testify/assert"
)

func TestNewLocal(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "rclonefs")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test configuration
	cfg := &RcloneLocalConfig{
		Name:    "test",
		Root:    tmpDir,
		BaseURL: "http://example.com",
	}

	// Call the NewLocal function
	disk, err := NewLocal(cfg)
	assert.NoError(t, err)

	// Assert that the returned disk is of type *RcloneLocal
	rl, ok := disk.(*RcloneLocal)
	assert.True(t, ok)

	// Assert that the disk name is set correctly
	assert.Equal(t, cfg.Name, rl.DiskName)

	// Assert that the root directory is created
	_, err = os.Stat(cfg.Root)
	assert.NoError(t, err)

	// Assert that the base URL is set correctly
	assert.Equal(t, cfg.BaseURL, rl.BaseURL)

	// Assert that the file system driver is created correctly
	fs, ok := rl.Fs.(*local.Fs)
	assert.True(t, ok)
	assert.Equal(t, rl.DiskName, fs.Name())
	assert.Equal(t, rl.Root, fs.Root())
}
func TestRcloneLocalURL(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "rclonefs")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test configuration
	cfg := &RcloneLocalConfig{
		Name:    "test",
		Root:    tmpDir,
		BaseURL: "http://example.com",
	}

	// Create a new RcloneLocal instance
	rl, err := NewLocal(cfg)
	assert.NoError(t, err)

	// Test the URL method with a file path
	filepath := "file.txt"
	expectedURL := "http://example.com/file.txt"
	actualURL := rl.URL(filepath)
	assert.Equal(t, expectedURL, actualURL)

	// Test the URL method with an empty file path
	emptyFilepath := ""
	expectedEmptyURL := "http://example.com/"
	actualEmptyURL := rl.URL(emptyFilepath)
	assert.Equal(t, expectedEmptyURL, actualEmptyURL)

	localDisk, ok := rl.(*RcloneLocal)
	assert.True(t, ok)

	localDisk.GetBaseURL = func() string {
		return "http://custom-url.com"
	}

	customURL := "custom-file.txt"
	expectedCustomURL := "http://custom-url.com/custom-file.txt"
	actualCustomURL := localDisk.URL(customURL)
	assert.Equal(t, expectedCustomURL, actualCustomURL)
}
