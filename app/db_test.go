package app_test

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/stretchr/testify/assert"
)

func TestDBConfigClone(t *testing.T) {
	c := &app.DBConfig{
		Driver:       "mysql",
		Name:         "mydb",
		Host:         "localhost",
		Port:         "3306",
		User:         "root",
		Pass:         "password",
		Logger:       nil,
		LogQueries:   true,
		MigrationDir: "/path/to/migrations",
	}

	clone := c.Clone()
	assert.Equal(t, c.Driver, clone.Driver)
	assert.Equal(t, c.Name, clone.Name)
	assert.Equal(t, c.Host, clone.Host)
	assert.Equal(t, c.Port, clone.Port)
	assert.Equal(t, c.User, clone.User)
	assert.Equal(t, c.Pass, clone.Pass)
	assert.Equal(t, c.Logger, clone.Logger)
	assert.Equal(t, c.LogQueries, clone.LogQueries)
	assert.Equal(t, c.MigrationDir, clone.MigrationDir)
}
