package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalProvider_UploadAndDownload(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalProvider(dir)

	ctx := context.Background()
	data := []byte("hello world")

	if err := p.UploadFile(ctx, "test-bucket", "file.txt", data); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	// Verify file on disk
	content, err := os.ReadFile(filepath.Join(dir, "test-bucket", "file.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("got %q, want %q", content, "hello world")
	}

	// GetDownloadURL should return a file:// URL
	url, err := p.GetDownloadURL(ctx, "test-bucket", "file.txt")
	if err != nil {
		t.Fatalf("GetDownloadURL: %v", err)
	}
	expected := "file://" + filepath.Join(dir, "test-bucket", "file.txt")
	if url != expected {
		t.Errorf("GetDownloadURL = %q, want %q", url, expected)
	}
}

func TestLocalProvider_DeleteFile(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalProvider(dir)

	ctx := context.Background()
	data := []byte("to be deleted")

	if err := p.UploadFile(ctx, "bucket", "del.txt", data); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if err := p.DeleteFile(ctx, "bucket", "del.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	// File should no longer exist
	_, err := os.Stat(filepath.Join(dir, "bucket", "del.txt"))
	if !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, got err: %v", err)
	}
}

func TestLocalProvider_DeleteNonexistent(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalProvider(dir)

	// Deleting a non-existent file should not error
	if err := p.DeleteFile(context.Background(), "bucket", "nope.txt"); err != nil {
		t.Fatalf("DeleteFile on non-existent file: %v", err)
	}
}

func TestLocalProvider_GetDownloadURL_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalProvider(dir)

	_, err := p.GetDownloadURL(context.Background(), "bucket", "missing.txt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// Compile-time interface compliance check.
var _ StorageProvider = (*LocalProvider)(nil)
