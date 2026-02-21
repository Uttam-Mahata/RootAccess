package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalProvider implements StorageProvider using the local file system.
type LocalProvider struct {
	basePath string
}

// NewLocalProvider creates a LocalProvider rooted at basePath.
func NewLocalProvider(basePath string) *LocalProvider {
	abs, _ := filepath.Abs(basePath)
	return &LocalProvider{basePath: abs}
}

// safePath returns the cleaned, absolute file path and ensures it stays
// within the provider's basePath to prevent path-traversal attacks.
func (p *LocalProvider) safePath(bucket, filename string) (string, error) {
	path := filepath.Join(p.basePath, bucket, filename)
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, p.basePath+string(filepath.Separator)) && path != p.basePath {
		return "", fmt.Errorf("path traversal detected")
	}
	return path, nil
}

func (p *LocalProvider) UploadFile(_ context.Context, bucket string, filename string, data []byte) error {
	path, err := p.safePath(bucket, filename)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create bucket directory: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (p *LocalProvider) DeleteFile(_ context.Context, bucket string, filename string) error {
	path, err := p.safePath(bucket, filename)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *LocalProvider) GetDownloadURL(_ context.Context, bucket string, filename string) (string, error) {
	path, err := p.safePath(bucket, filename)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}
	return "file://" + path, nil
}
