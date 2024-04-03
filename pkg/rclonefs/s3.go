package rclonefs

import (
	"context"

	"github.com/fastschema/fastschema/app"
	"github.com/rclone/rclone/backend/s3"
	rclonefs "github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
)

type RcloneS3 struct {
	*BaseRcloneDisk
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

func NewS3(cfg *RcloneS3Config) app.Disk {
	if cfg.ChunkSize < rclonefs.SizeSuffix(1024*1024*5) {
		cfg.ChunkSize = rclonefs.SizeSuffix(1024 * 1024 * 5)
	}

	rs3 := &RcloneS3{
		BaseRcloneDisk: &BaseRcloneDisk{
			DiskName: cfg.Name,
			Root:     cfg.Root,
		},
		Root:            cfg.Root,
		Provider:        cfg.Provider,
		Region:          cfg.Region,
		Endpoint:        cfg.Endpoint,
		ChunkSize:       cfg.ChunkSize,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		BaseURL:         cfg.BaseURL,
		ACL:             cfg.ACL,
	}

	rs3.BaseRcloneDisk.GetURL = rs3.URL
	cfgMap := &configmap.Simple{}
	cfgMap.Set("provider", cfg.Provider)
	cfgMap.Set("bucket", cfg.Bucket)
	cfgMap.Set("region", cfg.Region)
	cfgMap.Set("endpoint", cfg.Endpoint)
	cfgMap.Set("chunk_size", cfg.ChunkSize.String())
	cfgMap.Set("access_key_id", cfg.AccessKeyID)
	cfgMap.Set("secret_access_key", cfg.SecretAccessKey)
	cfgMap.Set("acl", cfg.ACL)
	cfgMap.Set("bucket_acl", cfg.ACL)

	fsDriver, err := s3.NewFs(context.Background(), "s3", cfg.Bucket, cfgMap)

	if err != nil {
		panic(err)
	}

	rs3.Fs = fsDriver

	return rs3
}

func (r *RcloneS3) URL(filepath string) string {
	return r.BaseURL + filepath
}

func (r *RcloneS3) Delete(ctx context.Context, filepath string) error {
	return nil
}
