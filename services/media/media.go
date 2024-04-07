package mediaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

type MediaService struct {
	app app.App
}

func NewMediaService(app app.App) *MediaService {
	return &MediaService{
		app: app,
	}
}

func (m *MediaService) MediaListHook(query *app.QueryOption, entities []*schema.Entity) ([]*schema.Entity, error) {
	if query.Model.Schema().Name != "media" {
		return entities, nil
	}

	for _, entity := range entities {
		path := entity.GetString("path")
		if path != "" {
			entity.Set("url", m.app.Disk().URL(entity.GetString("path")))
		}
	}

	return entities, nil
}
