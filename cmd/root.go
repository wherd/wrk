package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wrk",
	Short: "Manage work",
	Long: `Utility application for managing your programming sessions.

    It streamlines your workflow with commands for switching branches,
    listing sessions, saving progress, identifying changed files,
    uploading to FTP, and committing changes.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		branchCommand,
		listSessionsCommand,
		saveSessionCommand,
		listFilesCommand)
}
