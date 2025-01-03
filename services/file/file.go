package file

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
)

type AppLike interface {
	DB() db.Client
	Disk(names ...string) fs.Disk
}

type FileService struct {
	DB   func() db.Client
	Disk func(names ...string) fs.Disk
}

func New(app AppLike) *FileService {
	return &FileService{
		DB:   app.DB,
		Disk: app.Disk,
	}
}

func (m *FileService) CreateResource(api *fs.Resource) {
	api.Group("file").
		Add(fs.NewResource("upload", m.Upload, &fs.Meta{Post: "/upload"})).
		Add(fs.NewResource("delete", m.Delete, &fs.Meta{Delete: "/"}))
}

func (m *FileService) FileListHook(
	ctx context.Context,
	query *db.QueryOption,
	entities []*entity.Entity,
) ([]*entity.Entity, error) {
	if query.Schema == nil {
		return entities, nil
	}

	if query.Schema.Name != "file" {
		return entities, nil
	}

	for _, entity := range entities {
		path := entity.GetString("path")
		disk := m.Disk(entity.GetString("disk"))

		if path != "" && disk != nil {
			entity.Set("url", disk.URL(path))
		}
	}

	return entities, nil
}
