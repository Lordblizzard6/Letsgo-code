package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ideCmd)
	ideCmd.Flags().String("path", ".", "Path to open in IDE")
}

var ideCmd = &cobra.Command{
	Use:   "ide",
	Short: "Open current project in IDE",
	Long:  `Open the current project in your configured IDE.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")

		var command string
		switch runtime.GOOS {
		case "darwin":
			command = "open"
		case "windows":
			command = "start"
		default:
			command = "xdg-open"
		}

		fmt.Printf("Opening %s in IDE...\n", path)
		exec.Command(command, path).Start()
	},
}
