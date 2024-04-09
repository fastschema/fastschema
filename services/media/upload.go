package mediaservice

import (
	"sync"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

func (m *MediaService) Upload(c app.Context, _ *any) (*app.Map, error) {
	files, err := c.Files()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.BadRequest("No files found")
	}

	var wg sync.WaitGroup
	wg.Add(len(files))
	uploadedFiles := make([]*app.File, 0)
	errorFiles := make([]*app.File, 0)

	for i, file := range files {
		go func(file *app.File, i int) {
			defer wg.Done()

			_, err := m.Disk().Put(c.Context(), file)
			if err != nil {
				c.Logger().Errorf("Error uploading file: %s", err)
				errorFiles = append(errorFiles, file)
			} else {
				savedFile, err := m.saveFile(file, c.User().ID)
				if err != nil {
					c.Logger().Errorf("Error saving file: %s", err)
					errorFiles = append(errorFiles, file)
				} else {
					uploadedFiles = append(uploadedFiles, savedFile)
				}
			}
		}(file, i)
	}

	wg.Wait()

	return &app.Map{
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
