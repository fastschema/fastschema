package file

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/google/uuid"
)

func (m *FileService) Delete(c fs.Context, fileIDs []uuid.UUID) (any, error) {
	conditions := []*db.Predicate{
		db.In("id", utils.Map(fileIDs, func(id uuid.UUID) any {
			return id
		})),
	}
	files, err := db.Builder[*fs.File](m.DB()).Where(conditions...).Get(c)
	if err != nil {
		return nil, errors.InternalServerError("Failed to get files: %s", err)
	}

	if _, err := db.Delete[*fs.File](c, m.DB(), conditions); err != nil {
		return nil, errors.InternalServerError("Failed to delete files: %s", err)
	}

	for _, file := range files {
		disk := m.Disk(file.Disk)
		if err := disk.Delete(c, file.Path); err != nil {
			return nil, errors.InternalServerError("Failed to delete file: %s", err)
		}
	}

	return nil, nil
}
