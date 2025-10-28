package main

import (
	"fmt"
	"os"

	"github.com/O6lvl4/syncenv/internal/cli"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "syncenv",
		Short: "Sync environment variables with cloud storage",
		Long: `syncenv - Version-controlled environment variable management

syncenv helps you manage environment variables across different versions
of your application by syncing them with cloud storage (AWS S3, Azure Blob,
or Google Cloud Storage).

It automatically detects the current Git tag or branch and stores/retrieves
environment variables accordingly, making it easy to maintain different
configurations for different versions of your application.`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(cli.NewInitCmd())
	rootCmd.AddCommand(cli.NewPushCmd())
	rootCmd.AddCommand(cli.NewPullCmd())
	rootCmd.AddCommand(cli.NewListCmd())
	rootCmd.AddCommand(cli.NewDiffCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
