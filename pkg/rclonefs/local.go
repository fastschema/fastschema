package rclonefs

import (
	"context"
	"net/url"
	"os"

	"github.com/fastschema/fastschema/fs"
	"github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/fs/config/configmap"
)

type RcloneLocalConfig struct {
	Name       string        `json:"name"`
	Root       string        `json:"root"`
	BaseURL    string        `json:"base_url"`
	PublicPath string        `json:"public_path"`
	GetBaseURL func() string `json:"-"`
}

type RcloneLocal struct {
	*fs.DiskBase
	*BaseRcloneDisk

	config *RcloneLocalConfig
}

func NewLocal(config *RcloneLocalConfig) (fs.Disk, error) {
	diskBase := &fs.DiskBase{
		DiskName: config.Name,
		Root:     config.Root,
	}

	rl := &RcloneLocal{
		config:   config,
		DiskBase: diskBase,
		BaseRcloneDisk: &BaseRcloneDisk{
			Disk:           config.Name,
			UploadFilePath: diskBase.UploadFilePath,
			IsAllowedMime:  diskBase.IsAllowedMime,
		},
	}

	rl.GetURL = rl.URL

	if err := os.MkdirAll(config.Root, os.ModePerm); err != nil {
		return nil, err
	}

	cfgMap := configmap.New()
	cfgMap.Set("root", rl.config.Root)
	fsDriver, err := local.NewFs(context.Background(), rl.Disk, rl.config.Root, cfgMap)

	if err != nil {
		return nil, err
	}

	rl.Fs = fsDriver

	return rl, nil
}

func (r *RcloneLocal) Root() string {
	return r.config.Root
}

func (r *RcloneLocal) LocalPublicPath() string {
	return r.config.PublicPath
}

func (r *RcloneLocal) URL(filepath string) string {
	baseURL := r.config.BaseURL
	if r.config.GetBaseURL != nil {
		baseURL = r.config.GetBaseURL()
	}
	url, _ := url.JoinPath(baseURL, r.config.PublicPath, filepath)
	return url
}
