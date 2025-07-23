package fs

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// filenameRemoveCharsRegexp is a regular expression used to sanitize filenames by removing
// characters that are not alphanumeric, underscores, hyphens, or dots.
var filenameRemoveCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
var dashRegexp = regexp.MustCompile(`\-+`)

// DiskBase is designed to be embedded in other disk-related structs that implement the fs.Disk interface.
// It provides a common set of functionalities for handling files, such as MIME type checking and generating upload paths.
//
// Fields:
//   - DiskName: A unique identifier for the disk, typically assigned from config.Name.
//   - Root: The root directory path where files managed by this disk are stored, typically assigned from config.Root.
//
// Methods:
//   - Name(): Returns the DiskName of the disk. This can be used to retrieve the identifier of the disk.
//   - IsAllowedMime(mime string): Checks if the provided MIME type is allowed for upload based on a predefined list of allowed file types.
//     Returns true if the MIME type is allowed, false otherwise.
//   - UploadFilePath(filename string): Generates a path for uploading a file based on the current time and the provided filename.
//     The path includes the year and month as directories, followed by a timestamp and the sanitized filename.
//
// Usage:
// DiskBase is intended to be used as a foundational component for managing file storage. It can be embedded in other structs
// to provide them with basic file handling capabilities. The GetURL function should be implemented to suit the specific needs
// of the application, allowing for flexible URL generation strategies.
//
// Example:
//
//	type CustomDisk struct {
//	    fs.DiskBase
//	    // Additional fields and methods specific to CustomDisk
//	}
type DiskBase struct {
	DiskName string `json:"name"`
	Root     string
}

func (r *DiskBase) Name() string {
	return r.DiskName
}

func (r *DiskBase) IsAllowedMime(mime string) bool {
	for _, allowedFileType := range AllowedFileTypes {
		allowedFileType = strings.Split(allowedFileType, ";")[0]
		if allowedFileType == mime {
			return true
		}
	}

	return false
}

func (r *DiskBase) UploadFilePath(filename string) string {
	now := time.Now()
	filename = filenameRemoveCharsRegexp.ReplaceAllString(filename, "-")
	filename = dashRegexp.ReplaceAllString(filename, "-")
	filename = strings.ReplaceAll(filename, "-.", ".")
	return path.Join(
		strconv.Itoa(now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%d_%s", now.UnixMicro(), filename),
	)
}
