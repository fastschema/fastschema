package fs

import (
	"context"
	"io"
	"mime/multipart"
	"time"
)

const TraceID = "trace_id"

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

var (
	ContextKeyTraceID = ContextKey(TraceID)
)

type Traceable interface {
	TraceID() string
}

// Disk is the interface that defines the methods that a disk must implement
type Disk interface {
	Name() string
	Root() string
	URL(filepath string) string
	Delete(c context.Context, filepath string) error
	Put(c context.Context, file *File) (*File, error)
	PutReader(c context.Context, in io.Reader, size uint64, mime, dst string) (*File, error)
	PutMultipart(c context.Context, m *multipart.FileHeader, dsts ...string) (*File, error)
	LocalPublicPath() string
}

// File holds the file data
type File struct {
	_         any        `json:"-" fs:"namespace=files;label_field=name"`
	ID        uint64     `json:"id,omitempty"`
	Disk      string     `json:"disk,omitempty"`
	Name      string     `json:"name,omitempty"`
	Path      string     `json:"path,omitempty"`
	Type      string     `json:"type,omitempty"`
	Size      uint64     `json:"size,omitempty"`
	UserID    uint64     `json:"user_id,omitempty"`
	User      *User      `json:"user,omitempty" fs.relation:"{'type':'o2m','schema':'user','field':'files','owner':false,'fk_columns':{'target_column':'user_id'}}"`
	URL       string     `json:"url,omitempty" fs:"-"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Reader    io.Reader  `json:"-"`
}

// DiskConfig holds the disk configuration
type DiskConfig struct {
	Name            string        `json:"name"`
	Driver          string        `json:"driver"`
	Root            string        `json:"root"`
	BaseURL         string        `json:"base_url"`
	PublicPath      string        `json:"public_path"`
	GetBaseURL      func() string `json:"-"`
	Provider        string        `json:"provider"`
	Endpoint        string        `json:"endpoint"`
	Region          string        `json:"region"`
	Bucket          string        `json:"bucket"`
	AccessKeyID     string        `json:"access_key_id"`
	SecretAccessKey string        `json:"secret_access_key"`
	ACL             string        `json:"acl"`
}

// Clone returns a clone of the disk configuration
func (dc *DiskConfig) Clone() *DiskConfig {
	return &DiskConfig{
		Name:            dc.Name,
		Driver:          dc.Driver,
		Root:            dc.Root,
		BaseURL:         dc.BaseURL,
		GetBaseURL:      dc.GetBaseURL,
		Provider:        dc.Provider,
		Endpoint:        dc.Endpoint,
		Region:          dc.Region,
		Bucket:          dc.Bucket,
		AccessKeyID:     dc.AccessKeyID,
		SecretAccessKey: dc.SecretAccessKey,
		ACL:             dc.ACL,
	}
}

// StorageConfig holds the storage configuration
type StorageConfig struct {
	DefaultDisk string        `json:"default_disk"`
	Disks       []*DiskConfig `json:"disks"`
}

// Clone returns a clone of the storage configuration
func (sc *StorageConfig) Clone() *StorageConfig {
	if sc == nil {
		return nil
	}

	clone := &StorageConfig{
		DefaultDisk: sc.DefaultDisk,
		Disks:       make([]*DiskConfig, len(sc.Disks)),
	}

	for i, dc := range sc.Disks {
		clone.Disks[i] = dc.Clone()
	}

	return clone
}

// AllowedFileTypes is a list of allowed file types
var AllowedFileTypes = []string{
	"text/xml",
	"text/xml; charset=utf-8",
	"text/plain",
	"text/plain; charset=utf-8",
	// "image/svg+xml",
	"image/jpeg",
	"image/pjpeg",
	"image/png",
	"image/gif",
	"image/x-icon",
	"application/pdf",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/powerpoint",
	"application/x-mspowerpoint",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"application/mspowerpoint",
	"application/vnd.ms-powerpoint",
	"application/vnd.openxmlformats-officedocument.presentationml.slideshow",
	"application/vnd.oasis.opendocument.text",
	"application/excel",
	"application/vnd.ms-excel",
	"application/x-excel",
	"application/x-msexcel",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	// "application/octet-stream",
	"audio/mpeg3",
	"audio/x-mpeg-3",
	"video/x-mpeg",
	"audio/m4a",
	"audio/ogg",
	"audio/wav",
	"audio/x-wav",
	"video/mp4",
	"video/x-m4v",
	"video/quicktime",
	"video/x-ms-asf",
	"video/x-ms-wmv",
	"application/x-troff-msvideo",
	"video/avi",
	"video/msvideo",
	"video/x-msvideo",
	"audio/mpeg",
	"video/mpeg",
	"video/ogg",
	"video/3gpp",
	"audio/3gpp",
	"video/3gpp2",
	"audio/3gpp2",
}
