package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var branchCommand = &cobra.Command{
	Use:   "br",
	Short: "Switch to branch",
	Long:  "Switches to given branch name, creates if not exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, err := exec.LookPath("git")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := exec.Command(path, "checkout", "-b", args[0]).Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
