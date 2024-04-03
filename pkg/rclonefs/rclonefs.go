package rclonefs

import (
	"path"

	"github.com/fastschema/fastschema/app"
)

func NewFromConfig(diskConfigs []*app.DiskConfig, localRoot string) []app.Disk {
	var disks []app.Disk

	for _, diskConfig := range diskConfigs {
		switch diskConfig.Driver {
		case "s3":
			disks = append(disks, NewS3(&RcloneS3Config{
				Name:            diskConfig.Name,
				Root:            diskConfig.Root,
				Provider:        diskConfig.Provider,
				Bucket:          diskConfig.Bucket,
				Region:          diskConfig.Region,
				Endpoint:        diskConfig.Endpoint,
				AccessKeyID:     diskConfig.AccessKeyID,
				SecretAccessKey: diskConfig.SecretAccessKey,
				BaseURL:         diskConfig.BaseURL,
				ACL:             diskConfig.ACL,
			}))
		case "local":
			disks = append(disks, NewLocal(&RcloneLocalConfig{
				Name:       diskConfig.Name,
				Root:       path.Join(localRoot, diskConfig.Root),
				BaseURL:    diskConfig.BaseURL,
				GetBaseURL: diskConfig.GetBaseURL,
			}))
		}
	}

	return disks
}
