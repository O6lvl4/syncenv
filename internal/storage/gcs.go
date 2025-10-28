package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/O6lvl4/syncenv/internal/config"
	"google.golang.org/api/iterator"
)

// GCSStorage implements Storage interface for Google Cloud Storage
type GCSStorage struct {
	client     *storage.Client
	bucketName string
	prefix     string
}

// NewGCSStorage creates a new GCS storage instance
func NewGCSStorage(cfg *config.Config) (*GCSStorage, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client:     client,
		bucketName: cfg.Storage.BucketName,
		prefix:     cfg.Storage.Prefix,
	}, nil
}

// Upload uploads data to GCS
func (g *GCSStorage) Upload(ctx context.Context, tag string, data []byte) error {
	objectName := BuildKey(g.prefix, tag)

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

// Download downloads data from GCS
func (g *GCSStorage) Download(ctx context.Context, tag string) ([]byte, error) {
	objectName := BuildKey(g.prefix, tag)

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectName)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %w", err)
	}

	return data, nil
}

// List returns all available tags from GCS
func (g *GCSStorage) List(ctx context.Context) ([]string, error) {
	var tags []string

	bucket := g.client.Bucket(g.bucketName)
	query := &storage.Query{Prefix: g.prefix}
	it := bucket.Objects(ctx, query)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list GCS objects: %w", err)
		}

		objectName := attrs.Name
		// Extract tag from object name (remove prefix and .env suffix)
		tag := objectName
		if g.prefix != "" && len(objectName) > len(g.prefix) {
			tag = objectName[len(g.prefix):]
		}
		if len(tag) > 4 && tag[len(tag)-4:] == ".env" {
			tag = tag[:len(tag)-4]
		}
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// Exists checks if a tag exists in GCS
func (g *GCSStorage) Exists(ctx context.Context, tag string) (bool, error) {
	objectName := BuildKey(g.prefix, tag)

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectName)
	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check GCS object: %w", err)
	}

	return true, nil
}

// Delete removes a tag from GCS
func (g *GCSStorage) Delete(ctx context.Context, tag string) error {
	objectName := BuildKey(g.prefix, tag)

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}
