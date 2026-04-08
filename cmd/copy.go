package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().Bool("last", true, "Copy last message")
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy last message to clipboard",
	Long:  `Copy the last message from Claude to the clipboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		session, _ := db.GetActiveSession()
		var sessionID string
		if session == nil {
			sessionID = "default"
		} else {
			sessionID = session.ID
		}

		history, err := db.GetHistory(sessionID)
		if err != nil || len(history) == 0 {
			fmt.Println("No messages to copy.")
			os.Exit(1)
		}

		lastMsg := history[len(history)-1]
		content := fmt.Sprintf("%v", lastMsg.Content)

		// Copy to clipboard
		var copyCmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			copyCmd = exec.Command("pbcopy")
		case "windows":
			copyCmd = exec.Command("clip")
		default:
			copyCmd = exec.Command("xclip", "-selection", "clipboard")
		}

		stdin, _ := copyCmd.StdinPipe()
		copyCmd.Start()
		stdin.Write([]byte(content))
		stdin.Close()
		copyCmd.Wait()

		fmt.Println("✓ Copied to clipboard")
	},
}
