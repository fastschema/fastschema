package rclonefs

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/stretchr/testify/assert"
)

func TestNewFromConfig(t *testing.T) {
	rootDir := t.TempDir()
	diskConfigs := []*app.DiskConfig{
		{
			Driver:          "s3",
			Name:            "s3-disk",
			Root:            "/root",
			Provider:        "DigitalOcean",
			Bucket:          "my-bucket",
			Region:          "us-west-2",
			Endpoint:        "https://s3.us-west-2.amazonaws.com",
			AccessKeyID:     "access-key",
			SecretAccessKey: "secret-key",
			BaseURL:         "https://example.com/s3",
			ACL:             "private",
		},
		{
			Driver:  "local",
			Name:    "local-disk",
			Root:    "",
			BaseURL: "https://example.com/local",
		},
	}

	disks, err := NewFromConfig(diskConfigs, rootDir)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(disks))

	// Check if the first disk is an S3 disk
	s3Disk, ok := disks[0].(*RcloneS3)
	assert.True(t, ok)
	assert.Equal(t, "s3-disk", s3Disk.Name())
	assert.Equal(t, "/root", s3Disk.Root)
	assert.Equal(t, "DigitalOcean", s3Disk.Provider)
	assert.Equal(t, "us-west-2", s3Disk.Region)
	assert.Equal(t, "https://s3.us-west-2.amazonaws.com", s3Disk.Endpoint)
	assert.Equal(t, "access-key", s3Disk.AccessKeyID)
	assert.Equal(t, "secret-key", s3Disk.SecretAccessKey)
	assert.Equal(t, "https://example.com/s3", s3Disk.BaseURL)
	assert.Equal(t, "private", s3Disk.ACL)

	// Check if the second disk is a local disk
	localDisk, ok := disks[1].(*RcloneLocal)
	assert.True(t, ok)
	assert.Equal(t, "local-disk", localDisk.Name())
	assert.Equal(t, rootDir, localDisk.Root)
	assert.Equal(t, "https://example.com/local", localDisk.BaseURL)
}
