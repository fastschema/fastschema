package rclonefs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewS3(t *testing.T) {
	config := &RcloneS3Config{
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

	disk, err := NewS3(config)

	assert.NoError(t, err)
	assert.NotNil(t, disk)

	rs3, ok := disk.(*RcloneS3)
	assert.True(t, ok)
	assert.Equal(t, config.Name, rs3.Disk)
	assert.Equal(t, config.Root, rs3.Root())
	assert.Equal(t, config.Provider, rs3.config.Provider)
	assert.Equal(t, config.Region, rs3.config.Region)
	assert.Equal(t, config.Endpoint, rs3.config.Endpoint)
	assert.Equal(t, config.ChunkSize, rs3.config.ChunkSize)
	assert.Equal(t, config.AccessKeyID, rs3.config.AccessKeyID)
	assert.Equal(t, config.SecretAccessKey, rs3.config.SecretAccessKey)
	assert.Equal(t, config.BaseURL, rs3.config.BaseURL)
	assert.Equal(t, config.ACL, rs3.config.ACL)
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
		BaseURL:         "http://base_url",
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
