package mediaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

func (m *MediaService) Upload(c app.Context, _ any) (_ app.Map, err error) {
	uploadedFiles := make([]*app.File, 0)
	errorFiles := make([]*app.File, 0)

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

		savedFile, err := m.saveFile(file, c.User().ID)
		if err != nil {
			c.Logger().Errorf("Error saving file: %s", err)
			errorFiles = append(errorFiles, file)
		} else {
			uploadedFiles = append(uploadedFiles, savedFile)
		}
	}

	return app.Map{
		"success": uploadedFiles,
		"error":   errorFiles,
	}, nil
}

func (m *MediaService) saveFile(file *app.File, userID uint64) (*app.File, error) {
	mediaModel, err := m.DB().Model("media")
	if err != nil {
		return nil, err
	}

	e := schema.NewEntity().
		Set("disk", file.Disk).
		Set("name", file.Name).
		Set("path", file.Path).
		Set("type", file.Type).
		Set("size", file.Size).
		Set("user_id", userID)

	file.ID, err = mediaModel.Mutation().Create(e)
	if err != nil {
		return nil, err
	}

	return file, nil
}
