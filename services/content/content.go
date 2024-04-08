package contentservice

import (
	"github.com/fastschema/fastschema/app"
)

type ContentServiceConfig interface {
	DB() app.DBClient
}

type ContentService struct {
	DB func() app.DBClient
}

func New(app ContentServiceConfig) *ContentService {
	return &ContentService{
		DB: app.DB,
	}
}
