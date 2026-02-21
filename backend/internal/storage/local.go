package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LocalProvider implements StorageProvider using the local file system.
type LocalProvider struct {
	basePath string
}

// NewLocalProvider creates a LocalProvider rooted at basePath.
func NewLocalProvider(basePath string) *LocalProvider {
	return &LocalProvider{basePath: basePath}
}

func (p *LocalProvider) UploadFile(_ context.Context, bucket string, filename string, data []byte) error {
	dir := filepath.Join(p.basePath, bucket)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create bucket directory: %w", err)
	}
	path := filepath.Join(dir, filename)
	return os.WriteFile(path, data, 0o644)
}

func (p *LocalProvider) DeleteFile(_ context.Context, bucket string, filename string) error {
	path := filepath.Join(p.basePath, bucket, filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *LocalProvider) GetDownloadURL(_ context.Context, bucket string, filename string) (string, error) {
	path := filepath.Join(p.basePath, bucket, filename)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}
	return "file://" + path, nil
}
