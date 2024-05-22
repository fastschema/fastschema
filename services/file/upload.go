package file

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

func (m *FileService) Upload(c fs.Context, _ any) (_ fs.Map, err error) {
	uploadedFiles := make([]*fs.File, 0)
	errorFiles := make([]*fs.File, 0)

	if m.Disk() == nil {
		return nil, errors.InternalServerError("Disk is not configured")
	}

	files, err := c.Files()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if _, err := m.Disk().Put(c.Context(), file); err != nil {
			c.Logger().Errorf("Error uploading file: %s", err)
			errorFiles = append(errorFiles, file)
			continue
		}

		savedFile, err := m.saveFile(c.Context(), file, c.User().ID)
		if err != nil {
			c.Logger().Errorf("Error saving file: %s", err)
			errorFiles = append(errorFiles, file)
		} else {
			uploadedFiles = append(uploadedFiles, savedFile)
		}
	}

	return fs.Map{
		"success": uploadedFiles,
		"error":   errorFiles,
	}, nil
}

func (m *FileService) saveFile(ctx context.Context, file *fs.File, userID uint64) (*fs.File, error) {
	return db.Create[*fs.File](ctx, m.DB(), schema.NewEntity().
		Set("disk", file.Disk).
		Set("name", file.Name).
		Set("path", file.Path).
		Set("type", file.Type).
		Set("size", file.Size).
		Set("user_id", userID))
}
