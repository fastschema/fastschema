package miniofs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/fastschema/fastschema/fs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// // filenameRemoveCharsRegexp is a regular expression used to sanitize filenames by removing
// // characters that are not alphanumeric, underscores, hyphens, or dots.
// var filenameRemoveCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
// var dashRegexp = regexp.MustCompile(`\-+`)

// MinioConfig holds the configuration for Minio storage.
type MinioConfig struct {
	// Name is a unique identifier for the disk configuration.
	// Example: "minio"
	Name string `json:"name"`

	// Root specifies the base directory within the bucket for storage.
	// Example: "uploads"
	Root string `json:"root"`

	// BaseURL is the base URL from which stored objects can be accessed directly.
	// Example: "https://minio.example.com"
	BaseURL string `json:"base_url"`

	// GetBaseURL (optional) is a function that dynamically provides the BaseURL.
	// This is useful for cases where the base URL may change at runtime.
	// Example: func() string { return "https://minio.example.com" }
	GetBaseURL func() string `json:"-"`

	// Bucket is the name of the bucket where objects are stored.
	// Example: "my-bucket"
	Bucket string `json:"bucket"`

	// AccessKeyID is the access key ID for authenticating with the Minio server.
	// Example: "my-access-key"
	AccessKeyID string `json:"access_key_id"`

	// SecretAccessKey is the secret access key for authenticating with the Minio server.
	// Example: "my-secret"
	SecretAccessKey string `json:"secret_access_key"`

	// UseSSL specifies whether to use SSL/TLS when connecting to the Minio server.
	UseSSL bool `json:"use_ssl"`
}

type MinioDisk struct {
	*fs.DiskBase
	Client *minio.Client
	config *MinioConfig
}

// NewMinioDisk creates a new MinioDisk instance with the provided configuration and SSL usage flag.
// It initializes the Minio client and returns the MinioDisk instance or an error if the client
// cannot be created.
func NewMinioDisk(diskConfig *fs.DiskConfig) (fs.Disk, error) {
	useSSL, err := extractUseSSLFromURL(diskConfig.BaseURL)
	if err != nil {
		return nil, err
	}

	config := &MinioConfig{
		Name:            diskConfig.Name,
		Root:            diskConfig.Root,
		BaseURL:         diskConfig.BaseURL,
		GetBaseURL:      diskConfig.GetBaseURL,
		Bucket:          diskConfig.Bucket,
		AccessKeyID:     diskConfig.AccessKeyID,
		SecretAccessKey: diskConfig.SecretAccessKey,
		UseSSL:          useSSL,
	}

	endpoint := removeSchemeFromURL(diskConfig.BaseURL) // Endpoint must be given without the scheme

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioDisk{
		Client: minioClient,
		config: config,
		DiskBase: &fs.DiskBase{
			DiskName: config.Name,
			Root:     config.Root,
		},
	}, nil
}

func (md *MinioDisk) Root() string {
	return md.config.Root
}

func (md *MinioDisk) Put(ctx context.Context, file *fs.File) (*fs.File, error) {
	if file.Path == "" {
		file.Path = md.UploadFilePath(file.Name)
	}

	newFile, err := md.PutReader(ctx, file.Reader, file.Size, file.Type, file.Path)
	if err != nil {
		return nil, err
	}

	file.Disk = newFile.Disk
	file.Size = newFile.Size
	file.URL = newFile.URL

	return file, nil
}

func (md *MinioDisk) PutReader(ctx context.Context, reader io.Reader, size uint64, fileType, dst string) (*fs.File, error) {
	info, err := md.Client.PutObject(ctx, md.config.Bucket, dst, reader, int64(size), minio.PutObjectOptions{ContentType: fileType})
	if err != nil {
		return nil, err
	}

	return &fs.File{
		Disk: md.config.Name,
		Path: dst,
		Type: fileType,
		Size: uint64(info.Size),
		URL:  md.URL(dst),
	}, nil
}

func (md *MinioDisk) PutMultipart(ctx context.Context, m *multipart.FileHeader, dsts ...string) (*fs.File, error) {
	f, err := m.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileHeader := make([]byte, 512)
	if _, err := f.Read(fileHeader); err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	fileType := http.DetectContentType(fileHeader)
	if !md.IsAllowedMime(strings.ToLower(fileType)) {
		return nil, errors.New("file type is not allowed")
	}

	dst := ""
	if len(dsts) > 0 {
		dst = dsts[0]
	} else {
		dst = md.UploadFilePath(m.Filename)
	}

	return md.PutReader(ctx, f, uint64(m.Size), fileType, dst)
}

func (md *MinioDisk) Delete(ctx context.Context, filepath string) error {
	return md.Client.RemoveObject(ctx, md.config.Bucket, filepath, minio.RemoveObjectOptions{})
}

func (md *MinioDisk) LocalPublicPath() string {
	// Implement based on your public path strategy
	return ""
}

// PresignedURL returns a presigned URL for the specified file path.
// The presigned URL allows access to the file for a limited time, even when the bucket is private.
// If an error occurs while generating the presigned URL, an empty string is returned.
func (md *MinioDisk) PresignedURL(filepath string) string {
	reqParams := make(url.Values)
	presignedURL, err := md.Client.PresignedGetObject(context.Background(), md.config.Bucket, filepath, time.Hour, reqParams)
	md.Client.GetObject(context.Background(), md.config.Bucket, filepath, minio.GetObjectOptions{})
	if err != nil {
		return ""
	}
	return presignedURL.String()
}

// URL returns the complete URL for the given filepath.
// If a base URL is configured, it appends the cleaned filepath to the base URL.
// Otherwise, it constructs the URL using the configured endpoint and bucket.
// Note: For the constructed URL to be accessible, the Minio bucket must allow public read access
// for the objects. If the bucket's objects are not publicly accessible, the URL will not permit
// access to the file.
//
// Parameters:
//   - filepath: The path to the file within the Minio bucket, relative to the root directory
//     configured for the MinioDisk instance.
//
// Returns:
//   - A string representing the complete URL to access the specified file. If an error occurs
//     during URL construction, an empty string is returned.
func (r *MinioDisk) URL(filepath string) string {
	if r.config.GetBaseURL != nil {
		baseURL := r.config.GetBaseURL()
		if baseURL[len(baseURL)-1] == '/' {
			baseURL = baseURL[:len(baseURL)-1]
		}
		cleanFilePath := path.Clean("/" + filepath)
		return baseURL + cleanFilePath
	}

	endpointURL, err := url.Parse(r.config.BaseURL)
	if err != nil {
		return ""
	}

	cleanPath := path.Join(r.config.Bucket, filepath)
	cleanedURL := fmt.Sprintf("%s://%s/%s", endpointURL.Scheme, endpointURL.Host, cleanPath)

	return cleanedURL
}

func (md *MinioDisk) GetURL(filepath string) string {
	return md.URL(filepath)
}

// ExtractUseSSLFromURL takes a URL as input and determines whether to use SSL
// based on the scheme of the URL. It returns true if the scheme is "https",
// indicating that SSL should be used, and false otherwise.
func extractUseSSLFromURL(inputURL string) (bool, error) {
	// Parse the input URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		// Return an error if the URL cannot be parsed
		return false, err
	}

	// Determine and return the useSSL flag based on the URL scheme
	return parsedURL.Scheme == "https", nil
}

// RemoveSchemeFromURL takes a URL string as input and returns the URL without its scheme (http or https).
// If the input URL does not have a scheme, it is returned unchanged.
func removeSchemeFromURL(inputURL string) string {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return inputURL
	}

	if parsedURL.Scheme != "" {
		return strings.TrimPrefix(parsedURL.String(), parsedURL.Scheme+"://")
	}

	return inputURL
}
