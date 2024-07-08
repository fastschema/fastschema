# miniofs

`miniofs` is a package that implements a MinIO storage interface for the fastschema headless CMS. It provides functionality to store and retrieve files using MinIO object storage, a high-performance, S3-compatible object storage solution.

## Features

- File upload and retrieval from MinIO
- URL generation for stored files
- Integration with fastschema's `fs.Disk` interface

## Configuration

To use `miniofs`, configure it with your MinIO settings. Here's an example of how to create a new MinIO disk:

```go
diskConfig := &fs.DiskConfig{
    Name:            "minio",
    Root:            "uploads",
    BaseURL:         "https://minio.example.com",
    Bucket:          "my-bucket",
    AccessKeyID:     "my-access-key",
    SecretAccessKey: "my-secret-key",
}

minioDisk, err := miniofs.NewMinioDisk(diskConfig)
if err != nil {
    // Handle error
}
```

## Important Security Considerations

1. **Public Read Access**: This package is designed to serve files that will be publicly accessible (e.g., for a blog or similar applications). Configure your MinIO bucket to allow public read access for stored objects.
2. **Bucket Creation**: Create the MinIO bucket manually and configure it for public read access, but not public write access.
3. **Public Nature of Files**: Files stored using this package will be publicly accessible. Avoid storing sensitive or private information without additional security measures.
4. **Write Access**: Only your application should have write permissions to the bucket.

## Usage

Basic example of using the `miniofs` package:

```go
ctx := context.Background()

file := &fs.File{
    Name:   "example.txt",
    Reader: strings.NewReader("Hello, World!"),
    Size:   13,
    Type:   "text/plain",
}

uploadedFile, err := minioDisk.Put(ctx, file)
if err != nil {
    // Handle error
}

fmt.Println("File uploaded:", uploadedFile.URL)
```

## Disclaimer

This package is designed for use cases where public read access to stored files is acceptable. For private or sensitive information, implement additional security measures or use a different storage solution.
