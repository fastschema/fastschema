package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestDiskConfigClone(t *testing.T) {
	dc := &fs.DiskConfig{
		Name:            "test",
		Driver:          "test",
		Root:            "test",
		BaseURL:         "test",
		GetBaseURL:      func() string { return "test" },
		Provider:        "test",
		Endpoint:        "test",
		Region:          "test",
		Bucket:          "test",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		ACL:             "test",
	}

	clone := dc.Clone()

	assert.Equal(t, dc.Name, clone.Name)
	assert.Equal(t, dc.Driver, clone.Driver)
	assert.Equal(t, dc.Root, clone.Root)
	assert.Equal(t, dc.BaseURL, clone.BaseURL)
	assert.Equal(t, dc.GetBaseURL(), clone.GetBaseURL())
	assert.Equal(t, dc.Provider, clone.Provider)
	assert.Equal(t, dc.Endpoint, clone.Endpoint)
	assert.Equal(t, dc.Region, clone.Region)
	assert.Equal(t, dc.Bucket, clone.Bucket)
	assert.Equal(t, dc.AccessKeyID, clone.AccessKeyID)
	assert.Equal(t, dc.SecretAccessKey, clone.SecretAccessKey)
	assert.Equal(t, dc.ACL, clone.ACL)
}

func TestStorageConfigCloneNil(t *testing.T) {
	var sc *fs.StorageConfig

	clone := sc.Clone()

	assert.Nil(t, clone)
}

func TestStorageConfigClone(t *testing.T) {
	sc := &fs.StorageConfig{
		DefaultDisk: "test",
		Disks: []*fs.DiskConfig{
			{
				Name:            "test1",
				Driver:          "test1",
				Root:            "test1",
				BaseURL:         "test1",
				GetBaseURL:      func() string { return "test1" },
				Provider:        "test1",
				Endpoint:        "test1",
				Region:          "test1",
				Bucket:          "test1",
				AccessKeyID:     "test1",
				SecretAccessKey: "test1",
				ACL:             "test1",
			},
			{
				Name:            "test2",
				Driver:          "test2",
				Root:            "test2",
				BaseURL:         "test2",
				GetBaseURL:      func() string { return "test2" },
				Provider:        "test2",
				Endpoint:        "test2",
				Region:          "test2",
				Bucket:          "test2",
				AccessKeyID:     "test2",
				SecretAccessKey: "test2",
				ACL:             "test2",
			},
		},
	}

	clone := sc.Clone()

	assert.Equal(t, sc.DefaultDisk, clone.DefaultDisk)
	assert.Equal(t, len(sc.Disks), len(clone.Disks))
	for i := range sc.Disks {
		assert.Equal(t, sc.Disks[i].Name, clone.Disks[i].Name)
		assert.Equal(t, sc.Disks[i].Driver, clone.Disks[i].Driver)
		assert.Equal(t, sc.Disks[i].Root, clone.Disks[i].Root)
		assert.Equal(t, sc.Disks[i].BaseURL, clone.Disks[i].BaseURL)
		assert.Equal(t, sc.Disks[i].GetBaseURL(), clone.Disks[i].GetBaseURL())
		assert.Equal(t, sc.Disks[i].Provider, clone.Disks[i].Provider)
		assert.Equal(t, sc.Disks[i].Endpoint, clone.Disks[i].Endpoint)
		assert.Equal(t, sc.Disks[i].Region, clone.Disks[i].Region)
		assert.Equal(t, sc.Disks[i].Bucket, clone.Disks[i].Bucket)
		assert.Equal(t, sc.Disks[i].AccessKeyID, clone.Disks[i].AccessKeyID)
		assert.Equal(t, sc.Disks[i].SecretAccessKey, clone.Disks[i].SecretAccessKey)
		assert.Equal(t, sc.Disks[i].ACL, clone.Disks[i].ACL)
	}
}
func TestContextKeyString(t *testing.T) {
	ck := fs.ContextKey("test")
	str := ck.String()

	assert.Equal(t, "test", str)
}
