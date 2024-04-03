package rclonefs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fastschema/fastschema/app"
	rclonefs "github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/object"
)

var filenameRemoveCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
var dashRegexp = regexp.MustCompile(`\-+`)

type BaseRcloneDisk struct {
	rclonefs.Fs
	DiskName string `json:"name"`
	Root     string
	GetURL   func(string) string
}

func (r *BaseRcloneDisk) Name() string {
	return r.DiskName
}

func (r *BaseRcloneDisk) Put(ctx context.Context, file *app.File) (*app.File, error) {
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
) (*app.File, error) {
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

	return &app.File{
		Disk: r.DiskName,
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
) (*app.File, error) {
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

func (r *BaseRcloneDisk) IsAllowedMime(mime string) bool {
	for _, allowedFileType := range app.AllowedFileTypes {
		allowedFileType = strings.Split(allowedFileType, ";")[0]
		if allowedFileType == mime {
			return true
		}
	}

	return false
}

func (r *BaseRcloneDisk) UploadFilePath(filename string) string {
	now := time.Now()
	filename = filenameRemoveCharsRegexp.ReplaceAllString(filename, "-")
	filename = dashRegexp.ReplaceAllString(filename, "-")
	filename = strings.ReplaceAll(filename, "-.", ".")
	return path.Join(
		r.Root,
		strconv.Itoa(now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%d_%s", now.UnixMicro(), filename),
	)
}

func (r *BaseRcloneDisk) Delete(ctx context.Context, filePath string) error {
	obj, err := r.Fs.NewObject(ctx, filePath)

	if err != nil {
		return err
	}

	return obj.Remove(ctx)
}
