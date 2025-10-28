package cli

import (
	"context"
	"fmt"

	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/O6lvl4/syncenv/internal/storage"
	"github.com/spf13/cobra"
)

// NewDiffCmd creates the diff command
func NewDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <tag1> <tag2>",
		Short: "Show differences between two environment versions",
		Long:  "Compare environment variables between two versions and display added, removed, and changed variables",
		Args:  cobra.ExactArgs(2),
		RunE:  runDiff,
	}

	return cmd
}

func runDiff(cmd *cobra.Command, args []string) error {
	tag1 := args[0]
	tag2 := args[1]

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

	ctx := context.Background()

	// Download first version
	fmt.Printf("Downloading %s...\n", tag1)
	data1, err := store.Download(ctx, tag1)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", tag1, err)
	}

	processedData1, err := processData(data1, cfg)
	if err != nil {
		return fmt.Errorf("failed to process %s: %w", tag1, err)
	}

	env1, err := parseDataToEnvMap(processedData1, cfg)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", tag1, err)
	}

	// Download second version
	fmt.Printf("Downloading %s...\n", tag2)
	data2, err := store.Download(ctx, tag2)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", tag2, err)
	}

	processedData2, err := processData(data2, cfg)
	if err != nil {
		return fmt.Errorf("failed to process %s: %w", tag2, err)
	}

	env2, err := parseDataToEnvMap(processedData2, cfg)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", tag2, err)
	}

	// Compare
	added, removed, changed := diffEnvMaps(env1, env2)

	// Display results
	fmt.Printf("\nDifferences between %s and %s:\n", tag1, tag2)
	fmt.Println("========================================")

	if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
		fmt.Println("No differences found.")
		return nil
	}

	if len(added) > 0 {
		fmt.Printf("\nAdded in %s:\n", tag2)
		for key, value := range added {
			fmt.Printf("  + %s=%s\n", key, value)
		}
	}

	if len(removed) > 0 {
		fmt.Printf("\nRemoved in %s:\n", tag2)
		for key, value := range removed {
			fmt.Printf("  - %s=%s\n", key, value)
		}
	}

	if len(changed) > 0 {
		fmt.Printf("\nChanged in %s:\n", tag2)
		for key, change := range changed {
			fmt.Printf("  ~ %s: %s\n", key, change)
		}
	}

	fmt.Printf("\nSummary: +%d -%d ~%d\n", len(added), len(removed), len(changed))

	return nil
}
