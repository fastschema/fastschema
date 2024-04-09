package contentservice

import (
	"github.com/fastschema/fastschema/app"
)

type AppLike interface {
	DB() app.DBClient
}

type ContentService struct {
	DB func() app.DBClient
}

func New(app AppLike) *ContentService {
	return &ContentService{
		DB: app.DB,
	}
}
