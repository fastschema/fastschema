package mediaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (m *MediaService) Delete(c app.Context, fileIDs *[]any) (any, error) {
	fileModel, err := m.app.DB().Model("media")
	if err != nil {
		return nil, errors.BadGateway("Failed to get model: %s", err)
	}

	fileEntities, err := fileModel.Query(db.In("id", *fileIDs)).Get(c.Context())
	if err != nil {
		return nil, errors.BadGateway("Failed to get files: %s", err)
	}

	fileMutation, err := fileModel.Mutation()
	if err != nil {
		return nil, errors.BadGateway("Failed to get mutation: %s", err)
	}

	if _, err = fileMutation.Where(db.In("id", *fileIDs)).Delete(); err != nil {
		return nil, errors.BadGateway("Failed to delete files: %s", err)
	}

	files := app.EntitiesToFiles(fileEntities)
	for _, file := range files {
		disk := m.app.Disk(file.Disk)
		if err := disk.Delete(c.Context(), file.Path); err != nil {
			return nil, errors.BadGateway("Failed to delete file: %s", err)
		}
	}

	return nil, nil
}
