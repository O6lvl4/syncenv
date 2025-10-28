package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = ".syncenv.yml"
)

// StorageType represents the cloud storage provider
type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeAzure StorageType = "azure"
	StorageTypeGCS   StorageType = "gcs"
)

// Config represents the syncenv configuration
type Config struct {
	Storage    StorageConfig    `yaml:"storage"`
	Encryption EncryptionConfig `yaml:"encryption"`
	EnvFile    string           `yaml:"env_file,omitempty"`    // Deprecated: use EnvFiles instead
	EnvFiles   []string         `yaml:"env_files,omitempty"`   // Multiple files support
}

// StorageConfig holds storage-specific configuration
type StorageConfig struct {
	Type StorageType `yaml:"type"`

	// Common
	Prefix string `yaml:"prefix,omitempty"`

	// AWS S3
	Bucket string `yaml:"bucket,omitempty"`
	Region string `yaml:"region,omitempty"`

	// Azure Blob Storage
	AccountName   string `yaml:"account_name,omitempty"`
	ContainerName string `yaml:"container_name,omitempty"`

	// Google Cloud Storage
	ProjectID  string `yaml:"project_id,omitempty"`
	BucketName string `yaml:"bucket_name,omitempty"`
}

// EncryptionConfig holds encryption settings
type EncryptionConfig struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `yaml:"key,omitempty"` // Hex-encoded encryption key (auto-generated)
}

// Load reads and parses the configuration file
func Load() (*Config, error) {
	configPath := filepath.Join(".", ConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if len(config.EnvFiles) == 0 && config.EnvFile == "" {
		config.EnvFile = ".env"
	}

	// Convert single env_file to env_files for unified handling
	if config.EnvFile != "" && len(config.EnvFiles) == 0 {
		config.EnvFiles = []string{config.EnvFile}
	}

	return &config, nil
}

// Save writes the configuration to file
func (c *Config) Save() error {
	configPath := filepath.Join(".", ConfigFileName)
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetEnvFiles returns the list of environment files to manage
func (c *Config) GetEnvFiles() []string {
	if len(c.EnvFiles) > 0 {
		return c.EnvFiles
	}
	if c.EnvFile != "" {
		return []string{c.EnvFile}
	}
	return []string{".env"}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	switch c.Storage.Type {
	case StorageTypeS3:
		if c.Storage.Bucket == "" {
			return fmt.Errorf("s3 bucket is required")
		}
		if c.Storage.Region == "" {
			return fmt.Errorf("s3 region is required")
		}
	case StorageTypeAzure:
		if c.Storage.AccountName == "" {
			return fmt.Errorf("azure account_name is required")
		}
		if c.Storage.ContainerName == "" {
			return fmt.Errorf("azure container_name is required")
		}
	case StorageTypeGCS:
		if c.Storage.BucketName == "" {
			return fmt.Errorf("gcs bucket_name is required")
		}
		if c.Storage.ProjectID == "" {
			return fmt.Errorf("gcs project_id is required")
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", c.Storage.Type)
	}

	return nil
}
