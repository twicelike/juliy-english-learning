package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a new LocalStorage service saving files under baseDir.
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

// Save saves data to a file under {baseDir}/{prefix}/{YYYY}/{MM}/{DD}/{filename}.
func (s *LocalStorage) Save(ctx context.Context, prefix string, filename string, mimeType string, data io.Reader) (*StoredFile, error) {
	now := time.Now()
	// Build key like: prefix/2023/10/25/filename
	keyPath := filepath.Join(prefix, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()))
	
	fullDir := filepath.Join(s.baseDir, keyPath)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", fullDir, err)
	}

	fullPath := filepath.Join(fullDir, filename)
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer file.Close()

	size, err := io.Copy(file, data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to %s: %w", fullPath, err)
	}

	// Use forward slashes for the key to be platform-independent
	key := fmt.Sprintf("%s/%04d/%02d/%02d/%s", prefix, now.Year(), now.Month(), now.Day(), filename)

	return &StoredFile{
		Key:      key,
		Path:     fullPath,
		MimeType: mimeType,
		Size:     size,
	}, nil
}

// Open opens a file by its key.
func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, *StoredFile, error) {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, err
	}

	// We don't store mime type explicitly in local filesystem unless we keep a DB mapping,
	// but the interface needs it. We can return an empty string or try to guess.
	// The DB will have the mime_type.
	stored := &StoredFile{
		Key:      key,
		Path:     fullPath,
		Size:     info.Size(),
	}

	return file, stored, nil
}

// Delete removes a file by its key.
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
