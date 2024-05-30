package file

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
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

func (m *FileService) FileListHook(query *db.QueryOption, entities []*schema.Entity) ([]*schema.Entity, error) {
	if query.Model.Schema().Name != "file" {
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
