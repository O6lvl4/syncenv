package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/O6lvl4/syncenv/internal/git"
	"github.com/O6lvl4/syncenv/internal/storage"
	"github.com/spf13/cobra"
)

// NewPullCmd creates the pull command
func NewPullCmd() *cobra.Command {
	var tag string
	var force bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull environment variables from cloud storage",
		Long:  "Download environment file from cloud storage for the current Git version or a specified tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(tag, force)
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Explicit tag to use (defaults to current Git tag/branch)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite local env file without confirmation")

	return cmd
}

func runPull(tagFlag string, force bool) error {
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

	// Create storage client
	store, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	// Check if tag exists
	ctx := context.Background()
	exists, err := store.Exists(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to check if tag exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("tag '%s' not found in storage. Run 'syncenv list' to see available versions", tag)
	}

	// Check if local files exist
	files := cfg.GetEnvFiles()
	if !force {
		existingFiles := []string{}
		for _, file := range files {
			if _, err := os.Stat(file); err == nil {
				existingFiles = append(existingFiles, file)
			}
		}

		if len(existingFiles) > 0 {
			if len(existingFiles) == 1 {
				fmt.Printf("WARNING: Local file '%s' already exists and will be overwritten.\n", existingFiles[0])
			} else {
				fmt.Printf("WARNING: %d local files already exist and will be overwritten.\n", len(existingFiles))
			}
			fmt.Print("Continue? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Pull cancelled.")
				return nil
			}
		}
	}

	// Download from storage
	fmt.Printf("Downloading from %s storage...\n", cfg.Storage.Type)
	data, err := store.Download(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Process data (decrypt if needed)
	if cfg.Encryption.Enabled {
		fmt.Println("Decrypting data...")
	}
	processedData, err := processData(data, cfg)
	if err != nil {
		return err
	}

	// Save to local files
	if len(files) == 1 {
		fmt.Printf("Writing to environment file: %s\n", files[0])
	} else {
		fmt.Printf("Extracting %d environment files...\n", len(files))
	}
	if err := saveEnvFiles(processedData, cfg); err != nil {
		return err
	}

	fmt.Printf("Successfully pulled environment variables with tag: %s\n", tag)
	return nil
}
