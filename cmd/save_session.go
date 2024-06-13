package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var saveSessionCommand = &cobra.Command{
	Use:   "sv",
	Short: "Save session",
	Long:  "Save working session for later",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, err := exec.LookPath("git")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := exec.Command(path, "stash", "push", "-uam", args[0]).Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
