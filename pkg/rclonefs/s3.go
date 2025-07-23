package rclonefs

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/rclone/rclone/backend/s3"
	rclonefs "github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
)

type RcloneS3Config struct {
	Name            string              `json:"name"`
	Root            string              `json:"root"`
	Provider        string              `json:"provider"`
	Bucket          string              `json:"bucket"`
	Region          string              `json:"region"`
	Endpoint        string              `json:"endpoint"`
	ChunkSize       rclonefs.SizeSuffix `json:"chunk_size"`
	CopyCutoff      rclonefs.SizeSuffix `json:"copy_cutoff"`
	AccessKeyID     string              `json:"access_key_id"`
	SecretAccessKey string              `json:"secret_access_key"`
	BaseURL         string              `json:"base_url"`
	ACL             string              `json:"acl"`
}

type RcloneS3 struct {
	*fs.DiskBase
	*BaseRcloneDisk

	config *RcloneS3Config
}

func NewS3(config *RcloneS3Config) (fs.Disk, error) {
	if config.ChunkSize < rclonefs.SizeSuffix(1024*1024*5) {
		config.ChunkSize = rclonefs.SizeSuffix(1024 * 1024 * 5)
	}

	if config.CopyCutoff < rclonefs.SizeSuffixBase {
		config.CopyCutoff = rclonefs.SizeSuffixBase
	}

	diskBase := &fs.DiskBase{
		DiskName: config.Name,
		Root:     config.Root,
	}

	rs3 := &RcloneS3{
		config:   config,
		DiskBase: diskBase,
		BaseRcloneDisk: &BaseRcloneDisk{
			Disk:           config.Name,
			UploadFilePath: diskBase.UploadFilePath,
			IsAllowedMime:  diskBase.IsAllowedMime,
		},
	}

	rs3.GetURL = rs3.URL
	cfgMap := &configmap.Simple{}
	cfgMap.Set("provider", config.Provider)
	cfgMap.Set("bucket", config.Bucket)
	cfgMap.Set("region", config.Region)
	cfgMap.Set("endpoint", config.Endpoint)
	cfgMap.Set("chunk_size", config.ChunkSize.String())
	cfgMap.Set("copy_cutoff", config.CopyCutoff.String())
	cfgMap.Set("access_key_id", config.AccessKeyID)
	cfgMap.Set("secret_access_key", config.SecretAccessKey)
	cfgMap.Set("acl", config.ACL)
	cfgMap.Set("bucket_acl", config.ACL)

	if config.Provider == "Minio" {
		cfgMap.Set("force_path_style", "true")
	}

	fsDriver, err := s3.NewFs(context.Background(), config.Name, config.Bucket, cfgMap)
	if err != nil {
		return nil, err
	}

	rs3.Fs = fsDriver

	return rs3, nil
}

func (r *RcloneS3) Root() string {
	return r.config.Root
}

func (r *RcloneS3) LocalPublicPath() string {
	return ""
}

func (r *RcloneS3) URL(filepath string) string {
	endpointURL, err := url.Parse(r.config.BaseURL)
	if err != nil {
		return ""
	}

	// Ensure the base URL does not end with a slash
	host := endpointURL.Host
	host = strings.TrimSuffix(host, "/")

	// Ensure the file path does not start with a slash
	cleanPath := path.Join(r.config.Bucket, filepath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	cleanedURL := fmt.Sprintf("%s://%s/%s", endpointURL.Scheme, host, cleanPath)

	return cleanedURL
}
