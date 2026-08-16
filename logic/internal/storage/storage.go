package storage

import (
	"context"
	"io"
)

// StoredFile represents metadata of a file saved in storage.
type StoredFile struct {
	Key      string
	Path     string
	MimeType string
	Size     int64
}

// StorageService defines the interface for object storage operations.
type StorageService interface {
	Save(ctx context.Context, prefix string, filename string, mimeType string, data io.Reader) (*StoredFile, error)
	Open(ctx context.Context, key string) (io.ReadCloser, *StoredFile, error)
	Delete(ctx context.Context, key string) error
}
