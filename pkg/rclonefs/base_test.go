package rclonefs

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/stretchr/testify/assert"
)

func TestBaseRcloneDiskName(t *testing.T) {
	r := &BaseRcloneDisk{
		DiskName: "mydisk",
	}

	name := r.Name()
	assert.Equal(t, "mydisk", name)
}
func TestBaseRcloneDiskPut(t *testing.T) {
	ctx := context.TODO()
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.txt"
	mockReader := bytes.NewReader([]byte("test"))
	cfgMap := configmap.New()
	cfgMap.Set("root", tmpDir)
	fsDriver, err := local.NewFs(context.Background(), "test_local_disk", tmpDir, cfgMap)
	assert.NoError(t, err)

	// Create a mock BaseRcloneDisk instance
	r := &BaseRcloneDisk{
		DiskName: "test_local_disk",
		Fs:       fsDriver,
		GetURL: func(path string) string {
			return "http://localhost:8080/" + path
		},
	}

	// Create a mock file1
	file1 := &fs.File{
		Path:   tmpFile,
		Name:   "test.txt",
		Reader: mockReader, // Provide a valid reader here
		Size:   uint64(mockReader.Len()),
		Type:   "text/plain",
	}

	newFile, err := r.Put(ctx, file1)
	assert.NoError(t, err)
	assert.Equal(t, "test_local_disk", newFile.Disk)
	assert.NotEmpty(t, newFile.Path)
	assert.NotEmpty(t, newFile.URL)

	// Create a mock file1
	file2 := &fs.File{
		Path:   "",
		Name:   "test.txt",
		Reader: mockReader, // Provide a valid reader here
		Size:   uint64(mockReader.Len()),
		Type:   "text/plain",
	}

	newFile2, err := r.Put(ctx, file2)
	assert.NoError(t, err)
	assert.Equal(t, "test_local_disk", newFile2.Disk)
	assert.NotEmpty(t, newFile2.Path)
}

func mockFileHeader(t *testing.T, tmpFilePath string) *multipart.FileHeader {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// create plain text temp file with some content
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill the image with a color (white in this case)
	c := color.RGBA{255, 255, 255, 255}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, c)
		}
	}

	f, err := os.Create(tmpFilePath)
	assert.NoError(t, err)
	defer f.Close()

	// Write the image to the file
	assert.NoError(t, png.Encode(f, img))

	// Open the file
	file, err := os.Open(tmpFilePath)
	assert.NoError(t, err)
	defer file.Close()

	// Create a form file
	fw, err := w.CreateFormFile("file", filepath.Base(file.Name()))
	assert.NoError(t, err)

	// Write file to the form file
	_, err = io.Copy(fw, file)
	assert.NoError(t, err)

	// Close the multipart writer so that the boundary is written
	assert.NoError(t, w.Close())

	// Create a multipart reader
	r := multipart.NewReader(&b, w.Boundary())

	// Read the multipart form
	form, err := r.ReadForm(10 << 20) // max memory 10MB
	assert.NoError(t, err)

	// Return the FileHeader of the file form-data
	return form.File["file"][0]
}

func TestBaseRcloneDiskPutMultipart(t *testing.T) {
	ctx := context.TODO()
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.png"
	cfgMap := configmap.New()
	cfgMap.Set("root", tmpDir)
	fsDriver, err := local.NewFs(context.Background(), "test_local_disk", tmpDir, cfgMap)
	assert.NoError(t, err)

	// Create a mock BaseRcloneDisk instance
	r := &BaseRcloneDisk{
		DiskName: "test_local_disk",
		Fs:       fsDriver,
		GetURL: func(path string) string {
			return "http://localhost:8080/" + path
		},
	}

	newFile, err := r.PutMultipart(ctx, mockFileHeader(t, tmpFile), tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, "test_local_disk", newFile.Disk)
	assert.NotEmpty(t, newFile.Path)
	assert.NotEmpty(t, newFile.URL)
}

func TestBaseRcloneDiskDelete(t *testing.T) {
	ctx := context.TODO()
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.png"
	cfgMap := configmap.New()
	cfgMap.Set("root", tmpDir)
	fsDriver, err := local.NewFs(context.Background(), "test_local_disk", tmpDir, cfgMap)
	assert.NoError(t, err)

	// Create a mock BaseRcloneDisk instance
	r := &BaseRcloneDisk{
		DiskName: "test_local_disk",
		Fs:       fsDriver,
		GetURL: func(path string) string {
			return "http://localhost:8080/" + path
		},
	}

	newFile, err := r.PutMultipart(ctx, mockFileHeader(t, tmpFile))
	assert.NoError(t, err)
	assert.Equal(t, "test_local_disk", newFile.Disk)
	assert.NotEmpty(t, newFile.Path)
	assert.NotEmpty(t, newFile.URL)

	err = r.Delete(ctx, newFile.Path)
	assert.NoError(t, err)
}
