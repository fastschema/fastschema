package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
)

func main() {
	ctx := context.Background()

	// Leave the default storage configuration
	// FastSchema will use the local storage at data/public as the default storage
	app := utils.Must(fastschema.New(&fs.Config{}))

	disk := app.Disk()
	file := utils.Must(disk.Put(ctx, &fs.File{
		Name:   "file.txt",
		Path:   "custom/file.txt",
		Type:   "text/plain",
		Size:   11,
		Reader: bytes.NewReader([]byte("Hello world")),
	}))
	fmt.Printf("File created: %s\n", spew.Sdump(file))

	// Create application with custom storage
	app = utils.Must(fastschema.New(&fs.Config{
		StorageConfig: &fs.StorageConfig{
			DefaultDisk: "local_public",
			DisksConfig: []*fs.DiskConfig{
				{
					Name:       "local_public",
					Driver:     "local",
					Root:       "./public",
					BaseURL:    "http://localhost:8000/files",
					PublicPath: "/files", // This will expose the files in the public path
				},
				{
					Name:   "local_private",
					Driver: "local",
					Root:   "./private",
				},
			},
		},
	}))

	// Create a file in the private storage
	privateDisk := app.Disk("local_private")
	file = utils.Must(privateDisk.Put(ctx, &fs.File{
		Name:   "private_file.txt",
		Path:   "custom/private_file.txt",
		Type:   "text/plain",
		Size:   11,
		Reader: bytes.NewReader([]byte("Hello world")),
	}))
	fmt.Printf("File created: %s\n", spew.Sdump(file))
}
