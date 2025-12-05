package cli

import (
	"os"

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
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
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
