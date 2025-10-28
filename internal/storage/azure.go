package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/O6lvl4/syncenv/internal/config"
)

// AzureStorage implements Storage interface for Azure Blob Storage
type AzureStorage struct {
	client        *azblob.Client
	containerName string
	prefix        string
}

// NewAzureStorage creates a new Azure Blob storage instance
func NewAzureStorage(cfg *config.Config) (*AzureStorage, error) {
	// Azure connection string should be set via AZURE_STORAGE_CONNECTION_STRING env var
	// or use DefaultAzureCredential for more authentication options
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connectionString == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_CONNECTION_STRING environment variable not set")
	}

	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}

	return &AzureStorage{
		client:        client,
		containerName: cfg.Storage.ContainerName,
		prefix:        cfg.Storage.Prefix,
	}, nil
}

// Upload uploads data to Azure Blob Storage
func (a *AzureStorage) Upload(ctx context.Context, tag string, data []byte) error {
	blobName := BuildKey(a.prefix, tag)

	_, err := a.client.UploadBuffer(ctx, a.containerName, blobName, data, nil)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure: %w", err)
	}

	return nil
}

// Download downloads data from Azure Blob Storage
func (a *AzureStorage) Download(ctx context.Context, tag string) ([]byte, error) {
	blobName := BuildKey(a.prefix, tag)

	resp, err := a.client.DownloadStream(ctx, a.containerName, blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from Azure: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Azure blob: %w", err)
	}

	return data, nil
}

// List returns all available tags from Azure Blob Storage
func (a *AzureStorage) List(ctx context.Context) ([]string, error) {
	var tags []string

	pager := a.client.NewListBlobsFlatPager(a.containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &a.prefix,
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Azure blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			blobName := *blob.Name
			// Extract tag from blob name (remove prefix and .env suffix)
			tag := blobName
			if a.prefix != "" && len(blobName) > len(a.prefix) {
				tag = blobName[len(a.prefix):]
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

// Exists checks if a tag exists in Azure Blob Storage
func (a *AzureStorage) Exists(ctx context.Context, tag string) (bool, error) {
	blobName := BuildKey(a.prefix, tag)

	_, err := a.client.DownloadStream(ctx, a.containerName, blobName, &azblob.DownloadStreamOptions{
		Range: azblob.HTTPRange{Count: 1},
	})
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Delete removes a tag from Azure Blob Storage
func (a *AzureStorage) Delete(ctx context.Context, tag string) error {
	blobName := BuildKey(a.prefix, tag)

	_, err := a.client.DeleteBlob(ctx, a.containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete from Azure: %w", err)
	}

	return nil
}
