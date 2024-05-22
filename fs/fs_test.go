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
		DisksConfig: []*fs.DiskConfig{
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
	assert.Equal(t, len(sc.DisksConfig), len(clone.DisksConfig))
	for i := range sc.DisksConfig {
		assert.Equal(t, sc.DisksConfig[i].Name, clone.DisksConfig[i].Name)
		assert.Equal(t, sc.DisksConfig[i].Driver, clone.DisksConfig[i].Driver)
		assert.Equal(t, sc.DisksConfig[i].Root, clone.DisksConfig[i].Root)
		assert.Equal(t, sc.DisksConfig[i].BaseURL, clone.DisksConfig[i].BaseURL)
		assert.Equal(t, sc.DisksConfig[i].GetBaseURL(), clone.DisksConfig[i].GetBaseURL())
		assert.Equal(t, sc.DisksConfig[i].Provider, clone.DisksConfig[i].Provider)
		assert.Equal(t, sc.DisksConfig[i].Endpoint, clone.DisksConfig[i].Endpoint)
		assert.Equal(t, sc.DisksConfig[i].Region, clone.DisksConfig[i].Region)
		assert.Equal(t, sc.DisksConfig[i].Bucket, clone.DisksConfig[i].Bucket)
		assert.Equal(t, sc.DisksConfig[i].AccessKeyID, clone.DisksConfig[i].AccessKeyID)
		assert.Equal(t, sc.DisksConfig[i].SecretAccessKey, clone.DisksConfig[i].SecretAccessKey)
		assert.Equal(t, sc.DisksConfig[i].ACL, clone.DisksConfig[i].ACL)
	}
}
