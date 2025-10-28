package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg *config.Config) (*S3Storage, error) {
	ctx := context.Background()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Storage.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Storage{
		client: client,
		bucket: cfg.Storage.Bucket,
		prefix: cfg.Storage.Prefix,
	}, nil
}

// Upload uploads data to S3
func (s *S3Storage) Upload(ctx context.Context, tag string, data []byte) error {
	key := BuildKey(s.prefix, tag)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// Download downloads data from S3
func (s *S3Storage) Download(ctx context.Context, tag string) ([]byte, error) {
	key := BuildKey(s.prefix, tag)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object: %w", err)
	}

	return data, nil
}

// List returns all available tags from S3
func (s *S3Storage) List(ctx context.Context) ([]string, error) {
	var tags []string

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			// Extract tag from key (remove prefix and .env suffix)
			tag := key
			if s.prefix != "" && len(key) > len(s.prefix) {
				tag = key[len(s.prefix):]
			}
			if len(tag) > 4 && tag[len(tag)-4:] == ".env" {
				tag = tag[:len(tag)-4]
			}
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	return tags, nil
}

// Exists checks if a tag exists in S3
func (s *S3Storage) Exists(ctx context.Context, tag string) (bool, error) {
	key := BuildKey(s.prefix, tag)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}

// Delete removes a tag from S3
func (s *S3Storage) Delete(ctx context.Context, tag string) error {
	key := BuildKey(s.prefix, tag)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}
