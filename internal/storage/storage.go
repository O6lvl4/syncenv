package storage

import (
	"context"
	"fmt"

	"github.com/O6lvl4/syncenv/internal/config"
)

// Storage defines the interface for cloud storage operations
type Storage interface {
	// Upload uploads data to the storage with the given tag
	Upload(ctx context.Context, tag string, data []byte) error

	// Download retrieves data from the storage for the given tag
	Download(ctx context.Context, tag string) ([]byte, error)

	// List returns all available tags
	List(ctx context.Context) ([]string, error)

	// Exists checks if a tag exists
	Exists(ctx context.Context, tag string) (bool, error)

	// Delete removes a tag from storage
	Delete(ctx context.Context, tag string) error
}

// New creates a new storage instance based on the configuration
func New(cfg *config.Config) (Storage, error) {
	switch cfg.Storage.Type {
	case config.StorageTypeS3:
		return NewS3Storage(cfg)
	case config.StorageTypeAzure:
		return NewAzureStorage(cfg)
	case config.StorageTypeGCS:
		return NewGCSStorage(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}

// BuildKey creates a storage key from a tag and optional prefix
func BuildKey(prefix, tag string) string {
	if prefix == "" {
		return fmt.Sprintf("%s.env", tag)
	}
	return fmt.Sprintf("%s%s.env", prefix, tag)
}
