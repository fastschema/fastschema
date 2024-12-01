# S3 Driver

The `s3` driver provides functionality to store and retrieve files using various object storage solutions, including MinIO, AWS S3, and more.

## MinIO Provider

Below is a description of how to use and configure the `Minio` provider.

### Features

- File upload and retrieval from multiple object storage providers
- URL generation for stored files
- Integration with fastschema's `fs.Disk` interface

### Configuration

To use the `Minio` provider, configure it with your storage settings. Here's an example of how to set up the `STORAGE` environment variable for a MinIO configuration:

```json
[
 {
 "name": "my_minio",
 "driver": "s3",
 "root": "/files",
 "provider": "Minio",
 "endpoint": "http://localhost:9000",
 "region": "",
 "bucket": "fastschematest",
 "access_key_id": "access_key_id",
 "secret_access_key": "secret_access_key",
 "base_url": "https://cdn.site.local"
 }
]
```

### Important Security Considerations

1. **Public Read Access**: This package is designed to serve files that will be publicly accessible (e.g., for a blog or similar applications). Configure your storage bucket to allow public read access for stored objects.
2. **Bucket Creation**: Create the storage bucket manually and configure it for public read access, but not public write access.
3. **Public Nature of Files**: Files stored using this package will be publicly accessible. Avoid storing sensitive or private information without additional security measures.
4. **Write Access**: Only your application should have write permissions to the bucket.

### Disclaimer

This package is designed for use cases where public read access to stored files is acceptable. For private or sensitive information, implement additional security measures or use a different storage solution.
