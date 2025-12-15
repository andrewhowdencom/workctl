package cli

import (
	"context"
	"fmt"
	"os"

	"workctl/internal/config"
	"workctl/internal/telemetry"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "workctl",
	Short: "A tool to batch assignments for agentic developers",
	Long: `workctl is a CLI tool designed to handle automatic assignments to an agentic developer (e.g. Google Jules).

It allows executing large chunks of work in series without checking in, enabling batch processing of tasks (e.g., overnight).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, &cfg.Otel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up telemetry: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down telemetry: %v\n", err)
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of workctl",
	Long:  `All software has versions. This is workctl's`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("workctl v0.1")
	},
}
