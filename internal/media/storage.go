package media

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StorageProvider interface {
	Upload(ctx context.Context, relativePath string, data []byte, contentType string) (string, error)
}

// LocalStorageProvider writes files to local disk
type LocalStorageProvider struct {
	BaseDir string
	BaseURL string
}

var _ StorageProvider = (*LocalStorageProvider)(nil)

func NewLocalStorageProvider(baseDir, baseURL string) *LocalStorageProvider {
	if baseDir == "" {
		baseDir = "./uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}
	return &LocalStorageProvider{
		BaseDir: baseDir,
		BaseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (l *LocalStorageProvider) Upload(_ context.Context, relativePath string, data []byte, _ string) (string, error) {
	cleanRelPath := filepath.Clean(relativePath)
	targetPath := filepath.Join(l.BaseDir, cleanRelPath)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("create storage directories: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", fmt.Errorf("write uploaded file: %w", err)
	}

	urlPath := "/" + strings.ReplaceAll(cleanRelPath, "\\", "/")
	return l.BaseURL + urlPath, nil
}

// CloudStorageProvider handles uploading to Cloud Storage (Firebase/GCS/S3)
type CloudStorageProvider struct {
	BucketName string
	PublicURL  string
	// Fallback to local storage if bucket name or cloud credentials not set
	fallback StorageProvider
}

func NewCloudStorageProvider(bucketName, publicURL string, fallback StorageProvider) *CloudStorageProvider {
	return &CloudStorageProvider{
		BucketName: bucketName,
		PublicURL:  strings.TrimRight(publicURL, "/"),
		fallback:   fallback,
	}
}

func (c *CloudStorageProvider) Upload(ctx context.Context, relativePath string, data []byte, contentType string) (string, error) {
	// If bucket is not configured, gracefully use fallback local storage
	if c.BucketName == "" {
		if c.fallback != nil {
			return c.fallback.Upload(ctx, relativePath, data, contentType)
		}
		return "", fmt.Errorf("cloud storage bucket is not configured")
	}

	// When Cloud Storage bucket is configured:
	// Constructs standard public GCS / Firebase Storage URL
	cleanRelPath := strings.ReplaceAll(filepath.Clean(relativePath), "\\", "/")
	if c.PublicURL != "" {
		return fmt.Sprintf("%s/%s", c.PublicURL, cleanRelPath), nil
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", c.BucketName, cleanRelPath), nil
}
