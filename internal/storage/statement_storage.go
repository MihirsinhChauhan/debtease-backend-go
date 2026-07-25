package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StatementStorage stores statement PDF blobs.
type StatementStorage interface {
	Put(ctx context.Context, key string, r io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

// LocalStatementStorage stores files on disk under a root directory.
type LocalStatementStorage struct {
	Root string
}

func NewLocalStatementStorage(root string) (*LocalStatementStorage, error) {
	if root == "" {
		root = "./data/statements"
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	return &LocalStatementStorage{Root: root}, nil
}

func (s *LocalStatementStorage) fullPath(key string) string {
	return filepath.Join(s.Root, filepath.FromSlash(key))
}

func (s *LocalStatementStorage) Put(ctx context.Context, key string, r io.Reader) error {
	_ = ctx
	path := s.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStatementStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	_ = ctx
	return os.Open(s.fullPath(key))
}

// ObjectKey builds a storage key for a user/debt/upload.
func ObjectKey(userID, debtID, uploadID string) string {
	return fmt.Sprintf("%s/%s/%s.pdf", userID, debtID, uploadID)
}
