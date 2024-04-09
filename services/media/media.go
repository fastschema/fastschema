package mediaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

type AppLike interface {
	DB() app.DBClient
	Disk(names ...string) app.Disk
}

type MediaService struct {
	DB   func() app.DBClient
	Disk func(names ...string) app.Disk
}

func New(app AppLike) *MediaService {
	return &MediaService{
		DB:   app.DB,
		Disk: app.Disk,
	}
}

func (m *MediaService) MediaListHook(query *app.QueryOption, entities []*schema.Entity) ([]*schema.Entity, error) {
	if query.Model.Schema().Name != "media" {
		return entities, nil
	}

	for _, entity := range entities {
		path := entity.GetString("path")
		disk := entity.GetString("disk")
		if path != "" {
			entity.Set("url", m.Disk(disk).URL(entity.GetString("path")))
		}
	}

	return entities, nil
}
