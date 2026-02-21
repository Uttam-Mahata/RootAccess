package storage

import "context"

// StorageProvider defines the interface for object storage operations.
type StorageProvider interface {
	UploadFile(ctx context.Context, bucket string, filename string, data []byte) error
	DeleteFile(ctx context.Context, bucket string, filename string) error
	GetDownloadURL(ctx context.Context, bucket string, filename string) (string, error)
}
