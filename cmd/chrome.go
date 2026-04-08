package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chromeCmd)
	chromeCmd.Flags().String("url", "", "URL to open in Chrome")
}

var chromeCmd = &cobra.Command{
	Use:   "chrome",
	Short: "Open Chrome browser",
	Long:  `Open Google Chrome browser with optional URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		if url == "" && len(args) > 0 {
			url = args[0]
		}

		var chromePath string
		switch runtime.GOOS {
		case "darwin":
			chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		case "windows":
			chromePath = `C:\Program Files\Google\Chrome\Application\chrome.exe`
		default:
			chromePath = "google-chrome"
		}

		if url != "" {
			fmt.Printf("Opening Chrome with: %s\n", url)
			exec.Command(chromePath, url).Start()
		} else {
			fmt.Println("Opening Chrome...")
			exec.Command(chromePath).Start()
		}
	},
}
