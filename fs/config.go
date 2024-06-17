package fs

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/logger"
)

type Config struct {
	Dir               string
	AppKey            string
	Port              string
	BaseURL           string
	DashURL           string
	APIBaseName       string
	DashBaseName      string
	Logger            logger.Logger
	LoggerConfig      *logger.Config // If Logger is set, LoggerConfig will be ignored
	DB                db.Client
	DBConfig          *db.Config // If DB is set, DBConfig will be ignored
	StorageConfig     *StorageConfig
	HideResourcesInfo bool
	SystemSchemas     []any // types to build the system schemas
	AuthConfig        *AuthConfig
}

func (ac *Config) Clone() *Config {
	c := &Config{
		Dir:               ac.Dir,
		AppKey:            ac.AppKey,
		Port:              ac.Port,
		BaseURL:           ac.BaseURL,
		DashURL:           ac.DashURL,
		APIBaseName:       ac.APIBaseName,
		DashBaseName:      ac.DashBaseName,
		Logger:            ac.Logger,
		DB:                ac.DB,
		HideResourcesInfo: ac.HideResourcesInfo,
		SystemSchemas:     append([]any{}, ac.SystemSchemas...),
		AuthConfig:        ac.AuthConfig.Clone(),
	}

	if ac.DBConfig != nil {
		c.DBConfig = ac.DBConfig.Clone()
	}

	if ac.StorageConfig != nil {
		c.StorageConfig = ac.StorageConfig.Clone()
	}

	return c
}
