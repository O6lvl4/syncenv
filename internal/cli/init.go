package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/O6lvl4/syncenv/internal/crypto"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize syncenv configuration",
		Long:  "Initialize syncenv by creating a configuration file and optionally generating an encryption key",
		RunE:  runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("syncenv initialization")
	fmt.Println("======================")
	fmt.Println()

	// Check if config already exists
	if _, err := os.Stat(config.ConfigFileName); err == nil {
		fmt.Printf("Configuration file %s already exists. Overwrite? (y/N): ", config.ConfigFileName)
		response, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			fmt.Println("Initialization cancelled.")
			return nil
		}
	}

	cfg := &config.Config{}

	// Storage type
	fmt.Println("Select storage type:")
	fmt.Println("1. AWS S3")
	fmt.Println("2. Azure Blob Storage")
	fmt.Println("3. Google Cloud Storage")
	fmt.Print("Choice (1-3): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		cfg.Storage.Type = config.StorageTypeS3
		fmt.Print("S3 Bucket name: ")
		bucket, _ := reader.ReadString('\n')
		cfg.Storage.Bucket = strings.TrimSpace(bucket)

		fmt.Print("AWS Region (e.g., us-west-2): ")
		region, _ := reader.ReadString('\n')
		cfg.Storage.Region = strings.TrimSpace(region)

	case "2":
		cfg.Storage.Type = config.StorageTypeAzure
		fmt.Print("Azure Storage Account name: ")
		account, _ := reader.ReadString('\n')
		cfg.Storage.AccountName = strings.TrimSpace(account)

		fmt.Print("Container name: ")
		container, _ := reader.ReadString('\n')
		cfg.Storage.ContainerName = strings.TrimSpace(container)

	case "3":
		cfg.Storage.Type = config.StorageTypeGCS
		fmt.Print("GCS Project ID: ")
		projectID, _ := reader.ReadString('\n')
		cfg.Storage.ProjectID = strings.TrimSpace(projectID)

		fmt.Print("GCS Bucket name: ")
		bucketName, _ := reader.ReadString('\n')
		cfg.Storage.BucketName = strings.TrimSpace(bucketName)

	default:
		return fmt.Errorf("invalid choice")
	}

	// Optional prefix
	fmt.Print("Storage path prefix (optional, press Enter to skip): ")
	prefix, _ := reader.ReadString('\n')
	cfg.Storage.Prefix = strings.TrimSpace(prefix)

	// Env file path(s)
	fmt.Print("Environment file path (default: .env, comma-separated for multiple): ")
	envFileInput, _ := reader.ReadString('\n')
	envFileInput = strings.TrimSpace(envFileInput)

	if envFileInput == "" {
		cfg.EnvFile = ".env"
	} else if strings.Contains(envFileInput, ",") {
		// Multiple files
		files := strings.Split(envFileInput, ",")
		for i, file := range files {
			files[i] = strings.TrimSpace(file)
		}
		cfg.EnvFiles = files
	} else {
		// Single file
		cfg.EnvFile = envFileInput
	}

	// Encryption
	fmt.Print("Enable encryption? (Y/n): ")
	encryptResponse, _ := reader.ReadString('\n')
	enableEncryption := !strings.HasPrefix(strings.ToLower(strings.TrimSpace(encryptResponse)), "n")
	cfg.Encryption.Enabled = enableEncryption

	if enableEncryption {
		// Generate a new encryption key and store it in the config
		key, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}

		cfg.Encryption.Key = crypto.EncodeKeyToString(key)
		fmt.Println("Encryption key generated and saved to configuration file.")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Printf("Configuration saved to %s\n", config.ConfigFileName)
	fmt.Println("syncenv is ready to use!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Run 'syncenv push' to upload your environment variables")
	fmt.Println("  - Run 'syncenv pull' to download environment variables")
	fmt.Println("  - Run 'syncenv list' to see all stored versions")

	return nil
}
