package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	// Create test config
	configContent := `storage:
  type: s3
  bucket: test-bucket
  region: us-west-2
  prefix: test/

encryption:
  enabled: true
  key: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

env_file: .env.test
`

	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify values
	if cfg.Storage.Type != StorageTypeS3 {
		t.Errorf("Expected storage type s3, got %s", cfg.Storage.Type)
	}
	if cfg.Storage.Bucket != "test-bucket" {
		t.Errorf("Expected bucket test-bucket, got %s", cfg.Storage.Bucket)
	}
	if cfg.Storage.Region != "us-west-2" {
		t.Errorf("Expected region us-west-2, got %s", cfg.Storage.Region)
	}
	if cfg.Storage.Prefix != "test/" {
		t.Errorf("Expected prefix test/, got %s", cfg.Storage.Prefix)
	}
	if !cfg.Encryption.Enabled {
		t.Error("Expected encryption enabled")
	}
	if cfg.Encryption.Key == "" {
		t.Error("Expected encryption key to be set")
	}
	if cfg.EnvFile != ".env.test" {
		t.Errorf("Expected env_file .env.test, got %s", cfg.EnvFile)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	// Create minimal config
	configContent := `storage:
  type: gcs
  project_id: test-project
  bucket_name: test-bucket

encryption:
  enabled: false
`

	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check defaults
	if cfg.EnvFile != ".env" {
		t.Errorf("Expected default env_file .env, got %s", cfg.EnvFile)
	}
}

func TestLoadConfigEncryptionDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	configContent := `storage:
  type: s3
  bucket: test-bucket
  region: us-west-2

encryption:
  enabled: true

env_file: .env
`

	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check that encryption is enabled (key should be empty until generated)
	if !cfg.Encryption.Enabled {
		t.Error("Expected encryption to be enabled")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing config file, got nil")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cfg := &Config{
		Storage: StorageConfig{
			Type:   StorageTypeS3,
			Bucket: "save-test-bucket",
			Region: "eu-west-1",
			Prefix: "saved/",
		},
		Encryption: EncryptionConfig{
			Enabled: true,
			Key:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
		EnvFile: ".env.prod",
	}

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}

	if loaded.Storage.Bucket != cfg.Storage.Bucket {
		t.Errorf("Bucket mismatch: expected %s, got %s", cfg.Storage.Bucket, loaded.Storage.Bucket)
	}
	if loaded.Storage.Region != cfg.Storage.Region {
		t.Errorf("Region mismatch: expected %s, got %s", cfg.Storage.Region, loaded.Storage.Region)
	}
	if loaded.EnvFile != cfg.EnvFile {
		t.Errorf("EnvFile mismatch: expected %s, got %s", cfg.EnvFile, loaded.EnvFile)
	}
}

func TestValidateS3Config(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Type:   StorageTypeS3,
			Bucket: "test-bucket",
			Region: "us-west-2",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Valid S3 config failed validation: %v", err)
	}
}

func TestValidateS3MissingBucket(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Type:   StorageTypeS3,
			Region: "us-west-2",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing S3 bucket, got nil")
	}
}

func TestValidateS3MissingRegion(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Type:   StorageTypeS3,
			Bucket: "test-bucket",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing S3 region, got nil")
	}
}

func TestValidateAzureConfig(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Type:          StorageTypeAzure,
			AccountName:   "testaccount",
			ContainerName: "testcontainer",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Valid Azure config failed validation: %v", err)
	}
}

func TestValidateAzureMissingFields(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "Missing account_name",
			config: Config{
				Storage: StorageConfig{
					Type:          StorageTypeAzure,
					ContainerName: "testcontainer",
				},
			},
		},
		{
			name: "Missing container_name",
			config: Config{
				Storage: StorageConfig{
					Type:        StorageTypeAzure,
					AccountName: "testaccount",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Error("Expected validation error, got nil")
			}
		})
	}
}

func TestValidateGCSConfig(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Type:       StorageTypeGCS,
			ProjectID:  "test-project",
			BucketName: "test-bucket",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Valid GCS config failed validation: %v", err)
	}
}

func TestValidateGCSMissingFields(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "Missing bucket_name",
			config: Config{
				Storage: StorageConfig{
					Type:      StorageTypeGCS,
					ProjectID: "test-project",
				},
			},
		},
		{
			name: "Missing project_id",
			config: Config{
				Storage: StorageConfig{
					Type:       StorageTypeGCS,
					BucketName: "test-bucket",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Error("Expected validation error, got nil")
			}
		})
	}
}

func TestValidateUnsupportedStorageType(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Type: "unsupported",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for unsupported storage type, got nil")
	}
}
