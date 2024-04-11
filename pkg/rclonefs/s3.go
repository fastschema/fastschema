package rclonefs

import (
	"context"

	"github.com/fastschema/fastschema/app"
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
	AccessKeyID     string              `json:"access_key_id"`
	SecretAccessKey string              `json:"secret_access_key"`
	BaseURL         string              `json:"base_url"`
	ACL             string              `json:"acl"`
}

type RcloneS3 struct {
	*BaseRcloneDisk
	config *RcloneS3Config
}

func NewS3(config *RcloneS3Config) (app.Disk, error) {
	if config.ChunkSize < rclonefs.SizeSuffix(1024*1024*5) {
		config.ChunkSize = rclonefs.SizeSuffix(1024 * 1024 * 5)
	}

	rs3 := &RcloneS3{
		config: config,
		BaseRcloneDisk: &BaseRcloneDisk{
			DiskName: config.Name,
			Root:     config.Root,
		},
	}

	rs3.BaseRcloneDisk.GetURL = rs3.URL
	cfgMap := &configmap.Simple{}
	cfgMap.Set("provider", config.Provider)
	cfgMap.Set("bucket", config.Bucket)
	cfgMap.Set("region", config.Region)
	cfgMap.Set("endpoint", config.Endpoint)
	cfgMap.Set("chunk_size", config.ChunkSize.String())
	cfgMap.Set("access_key_id", config.AccessKeyID)
	cfgMap.Set("secret_access_key", config.SecretAccessKey)
	cfgMap.Set("acl", config.ACL)
	cfgMap.Set("bucket_acl", config.ACL)

	fsDriver, err := s3.NewFs(context.Background(), "s3", config.Bucket, cfgMap)

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
	return r.config.BaseURL + filepath
}
