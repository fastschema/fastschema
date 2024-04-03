package contentservice

import "github.com/fastschema/fastschema/app"

type ContentService struct {
	app app.App
}

func NewContentService(
	app app.App,
) *ContentService {
	return &ContentService{
		app: app,
	}
}
