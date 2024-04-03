package rclonefs

import (
	"context"
	"os"

	"github.com/fastschema/fastschema/app"
	"github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/fs/config/configmap"
)

type RcloneLocal struct {
	*BaseRcloneDisk
	Root       string        `json:"root"`
	BaseURL    string        `json:"base_url"`
	GetBaseURL func() string `json:"-"`
}

type RcloneLocalConfig struct {
	Name       string        `json:"name"`
	Root       string        `json:"root"`
	BaseURL    string        `json:"base_url"`
	GetBaseURL func() string `json:"-"`
}

func NewLocal(cfg *RcloneLocalConfig) app.Disk {
	rl := &RcloneLocal{
		BaseRcloneDisk: &BaseRcloneDisk{
			DiskName: cfg.Name,
		},
		Root:       cfg.Root,
		BaseURL:    cfg.BaseURL,
		GetBaseURL: cfg.GetBaseURL,
	}

	rl.BaseRcloneDisk.GetURL = rl.URL

	if err := os.MkdirAll(cfg.Root, os.ModePerm); err != nil {
		panic(err)
	}

	cfgMap := configmap.New()
	cfgMap.Set("root", rl.Root)
	fsDriver, err := local.NewFs(context.Background(), rl.DiskName, rl.Root, cfgMap)

	if err != nil {
		panic(err)
	}

	rl.Fs = fsDriver

	return rl
}

func (r *RcloneLocal) URL(filepath string) string {
	if r.GetBaseURL != nil {
		return r.GetBaseURL() + "/" + filepath
	}
	return r.BaseURL + "/" + filepath
}
