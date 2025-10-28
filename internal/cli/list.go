package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/O6lvl4/syncenv/internal/git"
	"github.com/O6lvl4/syncenv/internal/storage"
	"github.com/spf13/cobra"
)

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all stored environment versions",
		Long:  "Display all available environment variable versions stored in cloud storage",
		RunE:  runList,
	}

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'syncenv init' first)", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create storage client
	store, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	// List all tags
	ctx := context.Background()
	fmt.Printf("Fetching list from %s storage...\n", cfg.Storage.Type)
	tags, err := store.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No versions found in storage.")
		return nil
	}

	// Sort tags (reverse order, newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	// Get current version if in a git repo
	var currentVersion string
	if git.IsGitRepository() {
		currentVersion, _ = git.GetCurrentVersion()
	}

	// Display tags
	fmt.Printf("\nAvailable versions (%d total):\n", len(tags))
	fmt.Println("========================================")
	for _, tag := range tags {
		marker := "  "
		if tag == currentVersion {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, tag)
	}

	if currentVersion != "" {
		fmt.Printf("\n* = current version (%s)\n", currentVersion)
	}

	return nil
}
