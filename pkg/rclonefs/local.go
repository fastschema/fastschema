package rclonefs

import (
	"context"
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
	*BaseRcloneDisk
	config *RcloneLocalConfig
}

func NewLocal(config *RcloneLocalConfig) (fs.Disk, error) {
	rl := &RcloneLocal{
		config: config,
		BaseRcloneDisk: &BaseRcloneDisk{
			DiskName: config.Name,
		},
	}

	rl.BaseRcloneDisk.GetURL = rl.URL

	if err := os.MkdirAll(config.Root, os.ModePerm); err != nil {
		return nil, err
	}

	cfgMap := configmap.New()
	cfgMap.Set("root", rl.config.Root)
	fsDriver, err := local.NewFs(context.Background(), rl.DiskName, rl.config.Root, cfgMap)

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
	if r.config.GetBaseURL != nil {
		return r.config.GetBaseURL() + "/" + filepath
	}
	return r.config.BaseURL + "/" + filepath
}
