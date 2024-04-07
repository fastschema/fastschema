package rclonefs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewS3(t *testing.T) {
	cfg := &RcloneS3Config{
		Name:            "s3",
		Root:            "/path/to/root",
		Provider:        "DigitalOcean",
		Region:          "region",
		Endpoint:        "endpoint",
		ChunkSize:       1024 * 1024,
		AccessKeyID:     "access_key_id",
		SecretAccessKey: "secret_access_key",
		BaseURL:         "base_url",
		ACL:             "acl",
	}

	disk, err := NewS3(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, disk)

	rs3, ok := disk.(*RcloneS3)
	assert.True(t, ok)
	assert.Equal(t, cfg.Name, rs3.DiskName)
	assert.Equal(t, cfg.Root, rs3.Root)
	assert.Equal(t, cfg.Root, rs3.BaseRcloneDisk.Root)
	assert.Equal(t, cfg.Provider, rs3.Provider)
	assert.Equal(t, cfg.Region, rs3.Region)
	assert.Equal(t, cfg.Endpoint, rs3.Endpoint)
	assert.Equal(t, cfg.ChunkSize, rs3.ChunkSize)
	assert.Equal(t, cfg.AccessKeyID, rs3.AccessKeyID)
	assert.Equal(t, cfg.SecretAccessKey, rs3.SecretAccessKey)
	assert.Equal(t, cfg.BaseURL, rs3.BaseURL)
	assert.Equal(t, cfg.ACL, rs3.ACL)
}

func TestRcloneS3URL(t *testing.T) {
	cfg := &RcloneS3Config{
		Name:            "s3",
		Root:            "/path/to/root",
		Provider:        "DigitalOcean",
		Region:          "region",
		Endpoint:        "endpoint",
		ChunkSize:       1024 * 1024,
		AccessKeyID:     "access_key_id",
		SecretAccessKey: "secret_access_key",
		BaseURL:         "base_url",
		ACL:             "acl",
	}

	disk, err := NewS3(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, disk)

	rs3, ok := disk.(*RcloneS3)
	assert.True(t, ok)

	filepath := "/path/to/file.txt"
	expectedURL := cfg.BaseURL + filepath
	actualURL := rs3.URL(filepath)

	assert.Equal(t, expectedURL, actualURL)
}
