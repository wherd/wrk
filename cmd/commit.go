package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var commitCommand = &cobra.Command{
	Use:   "cm",
	Short: "Commit",
	Long:  "Commit and push staged chages",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, err := exec.LookPath("git")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		args = append([]string{"commit", "-m"}, args...)

		if err := exec.Command(path, args...).Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := exec.Command(path, "push").Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
