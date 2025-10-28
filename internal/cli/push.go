package cli

import (
	"context"
	"fmt"

	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/O6lvl4/syncenv/internal/git"
	"github.com/O6lvl4/syncenv/internal/storage"
	"github.com/spf13/cobra"
)

// NewPushCmd creates the push command
func NewPushCmd() *cobra.Command {
	var tag string

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push environment variables to cloud storage",
		Long:  "Upload the local environment file to cloud storage, tagged with the current Git version or a specified tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(tag)
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Explicit tag to use (defaults to current Git tag/branch)")

	return cmd
}

func runPush(tagFlag string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'syncenv init' first)", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Determine tag
	var tag string
	if tagFlag != "" {
		tag = tagFlag
		fmt.Printf("Using explicit tag: %s\n", tag)
	} else {
		// Auto-detect from Git
		if !git.IsGitRepository() {
			return fmt.Errorf("not a git repository and no --tag specified")
		}

		tag, err = git.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("failed to determine Git version: %w (use --tag to specify manually)", err)
		}
		fmt.Printf("Auto-detected version from Git: %s\n", tag)
	}

	// Load env files
	files := cfg.GetEnvFiles()
	if len(files) == 1 {
		fmt.Printf("Reading environment file: %s\n", files[0])
	} else {
		fmt.Printf("Reading %d environment files...\n", len(files))
	}
	data, err := loadEnvFiles(cfg)
	if err != nil {
		return err
	}

	// Prepare data (encrypt if needed)
	if cfg.Encryption.Enabled {
		fmt.Println("Encrypting data...")
	}
	preparedData, err := prepareData(data, cfg)
	if err != nil {
		return err
	}

	// Create storage client
	store, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	// Check if tag already exists
	ctx := context.Background()
	exists, err := store.Exists(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to check if tag exists: %w", err)
	}

	if exists {
		fmt.Printf("WARNING: Tag '%s' already exists in storage. This will overwrite the existing version.\n", tag)
	}

	// Upload to storage
	fmt.Printf("Uploading to %s storage...\n", cfg.Storage.Type)
	if err := store.Upload(ctx, tag, preparedData); err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	fmt.Printf("Successfully pushed environment variables with tag: %s\n", tag)
	return nil
}
