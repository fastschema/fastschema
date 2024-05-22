package contentservice

import "github.com/fastschema/fastschema/db"

type AppLike interface {
	DB() db.Client
}

type ContentService struct {
	DB func() db.Client
}

func New(app AppLike) *ContentService {
	return &ContentService{
		DB: app.DB,
	}
}
