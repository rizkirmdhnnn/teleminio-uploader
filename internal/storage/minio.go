package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rizkirmdhnnn/teleminio-uploader/internal/config"
)

// MinioClient represents a MinIO client instance
type MinioClient struct {
	Client     *minio.Client
	BucketName string
}

// NewMinio initializes a new MinIO client
func NewMinio(cfg config.Config) (*MinioClient, error) {
	// Validate MinIO configuration
	if cfg.MinioHost == "" || cfg.MinioAccessKey == "" || cfg.MinioSecretKey == "" || cfg.MinioBucket == "" {
		return nil, fmt.Errorf("missing required MinIO configuration")
	}

	// Construct endpoint with host and port if port is provided
	endpoint := cfg.MinioHost
	if cfg.MinioPort != "" {
		endpoint = fmt.Sprintf("%s:%s", cfg.MinioHost, cfg.MinioPort)
	}

	// Use custom endpoint if provided
	if cfg.MinioEndpoint != "" {
		endpoint = cfg.MinioEndpoint
	}

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioSSL,
		Region: cfg.MinioRegion,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Create MinioClient instance
	minioClient := &MinioClient{
		Client:     client,
		BucketName: cfg.MinioBucket,
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return minioClient, nil
}

// UploadFile uploads a file to MinIO
func (m *MinioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// For large files (> 50MB), use multipart upload
	if size > 50*1024*1024 {
		return m.uploadLargeFile(ctx, objectName, reader, size, contentType)
	}

	// For smaller files, use regular upload
	_, err := m.Client.PutObject(ctx, m.BucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate a presigned URL for the uploaded object
	presignedURL, err := m.GetFileURL(ctx, objectName, time.Hour*24*7) // URL valid for 7 days
	if err != nil {
		return "", fmt.Errorf("file uploaded but failed to generate URL: %w", err)
	}

	return presignedURL, nil
}

// uploadLargeFile handles large file uploads using concurrent multipart upload
func (m *MinioClient) uploadLargeFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// Use PutObject with optimized settings for large files
	opts := minio.PutObjectOptions{
		ContentType: contentType,
		// Set part size to 5MB for better performance
		PartSize: 5 * 1024 * 1024,
	}

	// Upload the file
	_, err := m.Client.PutObject(ctx, m.BucketName, objectName, reader, size, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload large file: %w", err)
	}

	// Generate a presigned URL for the uploaded object
	presignedURL, err := m.GetFileURL(ctx, objectName, time.Hour*24*7) // URL valid for 7 days
	if err != nil {
		return "", fmt.Errorf("file uploaded but failed to generate URL: %w", err)
	}

	return presignedURL, nil
}

// GetFileURL generates a presigned URL for accessing a file
func (m *MinioClient) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	// Generate presigned URL
	reqParams := make(url.Values)
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.BucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// DownloadFile downloads a file from MinIO
func (m *MinioClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	// Get object
	object, err := m.Client.GetObject(ctx, m.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

// ListFiles lists all files in the bucket with an optional prefix
func (m *MinioClient) ListFiles(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	// Create a channel to receive objects
	objectCh := m.Client.ListObjects(ctx, m.BucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	// Iterate through the channel and collect objects
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// DeleteFile deletes a file from MinIO
func (m *MinioClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.Client.RemoveObject(ctx, m.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetObjectInfo gets information about an object
func (m *MinioClient) GetObjectInfo(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	info, err := m.Client.StatObject(ctx, m.BucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get object info: %w", err)
	}

	return info, nil
}

// GenerateObjectName generates a unique object name based on the original filename
func (m *MinioClient) GenerateObjectName(originalFilename string) string {
	// Extract file extension
	ext := filepath.Ext(originalFilename)
	baseName := originalFilename[:len(originalFilename)-len(ext)]

	// Generate a unique name using timestamp
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)
}
