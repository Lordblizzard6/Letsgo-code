package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clearCmd)
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the terminal screen",
	Long:  `Clear the terminal screen to provide a clean workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cmdClear *exec.Cmd
		if runtime.GOOS == "windows" {
			cmdClear = exec.Command("cmd", "/c", "cls")
		} else {
			cmdClear = exec.Command("clear")
		}
		cmdClear.Stdout = os.Stdout
		cmdClear.Stderr = os.Stderr
		if err := cmdClear.Run(); err != nil {
			// Fallback: print newlines
			for i := 0; i < 50; i++ {
				fmt.Println()
			}
		}
	},
}
