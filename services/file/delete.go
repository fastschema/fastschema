package file

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (m *FileService) Delete(c fs.Context, fileIDs []uint64) (any, error) {
	condition := db.In("id", utils.Map(fileIDs, func(id uint64) any {
		return id
	}))
	files, err := db.Query[*fs.File](m.DB()).Where(condition).Get(c.Context())
	if err != nil {
		return nil, errors.InternalServerError("Failed to get files: %s", err)
	}

	if _, err := db.Delete[*fs.File](c.Context(), m.DB(), condition); err != nil {
		return nil, errors.InternalServerError("Failed to delete files: %s", err)
	}

	for _, file := range files {
		disk := m.Disk(file.Disk)
		if err := disk.Delete(c.Context(), file.Path); err != nil {
			return nil, errors.InternalServerError("Failed to delete file: %s", err)
		}
	}

	return nil, nil
}
