package rclonefs

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/fastschema/fastschema/fs"
	rclonefs "github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/object"
)

type BaseRcloneDisk struct {
	rclonefs.Fs

	Disk           string `json:"name"`
	GetURL         func(string) string
	UploadFilePath func(string) string
	IsAllowedMime  func(string) bool
}

func (r *BaseRcloneDisk) Put(ctx context.Context, file *fs.File) (*fs.File, error) {
	if file.Path == "" {
		file.Path = r.UploadFilePath(file.Name)
	}

	newFile, err := r.PutReader(ctx, file.Reader, file.Size, file.Type, file.Path)
	if err != nil {
		return nil, err
	}

	file.Disk = newFile.Disk
	file.Size = newFile.Size
	file.URL = newFile.URL

	return file, nil
}

func (r *BaseRcloneDisk) PutReader(
	ctx context.Context,
	reader io.Reader,
	size uint64,
	fileType,
	dst string,
) (*fs.File, error) {
	objectInfo := object.NewStaticObjectInfo(
		dst,
		time.Now(),
		int64(size),
		true,
		nil,
		nil,
	)

	rs, err := r.Fs.Put(ctx, reader, objectInfo)

	if err != nil {
		return nil, err
	}

	return &fs.File{
		Disk: r.Disk,
		Path: dst,
		Type: fileType,
		Size: uint64(rs.Size()),
		URL:  r.GetURL(dst),
	}, nil
}

func (r *BaseRcloneDisk) PutMultipart(
	ctx context.Context,
	m *multipart.FileHeader,
	dsts ...string,
) (*fs.File, error) {
	f, err := m.Open()

	if err != nil {
		return nil, err
	}

	fileHeader := make([]byte, 512)

	if _, err := f.Read(fileHeader); err != nil {
		return nil, err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	dst := ""
	fileType := http.DetectContentType(fileHeader)

	if !r.IsAllowedMime(strings.ToLower(fileType)) {
		return nil, errors.New("file type is not allowed")
	}

	if len(dsts) > 0 {
		dst = dsts[0]
	} else {
		dst = r.UploadFilePath(m.Filename)
	}

	return r.PutReader(ctx, f, uint64(m.Size), fileType, dst)
}

func (r *BaseRcloneDisk) Delete(ctx context.Context, filePath string) error {
	obj, err := r.NewObject(ctx, filePath)

	if err != nil {
		return err
	}

	return obj.Remove(ctx)
}
