package mediaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (m *MediaService) Delete(c app.Context, fileIDs *[]any) (any, error) {
	fileModel, err := m.DB().Model("media")
	if err != nil {
		return nil, errors.InternalServerError("Failed to get model: %s", err)
	}

	fileEntities, err := fileModel.Query(app.In("id", *fileIDs)).Get(c.Context())
	if err != nil {
		return nil, errors.InternalServerError("Failed to get files: %s", err)
	}

	if _, err = fileModel.Mutation().Where(app.In("id", *fileIDs)).Delete(); err != nil {
		return nil, errors.InternalServerError("Failed to delete files: %s", err)
	}

	files := app.EntitiesToFiles(fileEntities)
	for _, file := range files {
		disk := m.Disk(file.Disk)
		if err := disk.Delete(c.Context(), file.Path); err != nil {
			return nil, errors.InternalServerError("Failed to delete file: %s", err)
		}
	}

	return nil, nil
}
