package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/stretchr/testify/assert"
)

func TestConfigClone(t *testing.T) {
	config := &fs.Config{
		Dir:          "/path/to/dir",
		AppKey:       "testKey",
		Port:         "8001",
		BaseURL:      "http://localhost:8001",
		DashURL:      "http://localhost:8001/testdash",
		APIBaseName:  "testapi",
		DashBaseName: "testdash",
		Logger:       logger.CreateMockLogger(true),
		DB:           nil,
		DBConfig:     &db.Config{},
		StorageConfig: &fs.StorageConfig{
			DefaultDisk: "local",
			DisksConfig: []*fs.DiskConfig{},
		},
		HideResourcesInfo: true,
		SystemSchemas:     []any{"schema1", "schema2"},
	}

	clone := config.Clone()

	assert.Equal(t, config.Dir, clone.Dir)
	assert.Equal(t, config.AppKey, clone.AppKey)
	assert.Equal(t, config.Port, clone.Port)
	assert.Equal(t, config.BaseURL, clone.BaseURL)
	assert.Equal(t, config.DashURL, clone.DashURL)
	assert.Equal(t, config.APIBaseName, clone.APIBaseName)
	assert.Equal(t, config.DashBaseName, clone.DashBaseName)
	assert.Equal(t, config.Logger, clone.Logger)
	assert.Equal(t, config.DB, clone.DB)
	assert.Equal(t, config.HideResourcesInfo, clone.HideResourcesInfo)
	assert.Equal(t, config.SystemSchemas, clone.SystemSchemas)

	// Ensure deep copy of DBConfig
	assert.Equal(t, config.DBConfig, clone.DBConfig)
	assert.Equal(t, config.DBConfig.Driver, clone.DBConfig.Driver)
	assert.Equal(t, config.DBConfig.Host, clone.DBConfig.Host)
	assert.Equal(t, config.DBConfig.Port, clone.DBConfig.Port)
	assert.Equal(t, config.DBConfig.User, clone.DBConfig.User)
	assert.Equal(t, config.DBConfig.Pass, clone.DBConfig.Pass)

	// Ensure deep copy of StorageConfig
	assert.Equal(t, config.StorageConfig, clone.StorageConfig)
	assert.Equal(t, config.StorageConfig.DefaultDisk, clone.StorageConfig.DefaultDisk)
	assert.Equal(t, config.StorageConfig.DisksConfig, clone.StorageConfig.DisksConfig)
}
