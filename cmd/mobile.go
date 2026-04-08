package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(mobileCmd)
}

var mobileCmd = &cobra.Command{
	Use:   "mobile",
	Short: "Show QR code for mobile access",
	Long:  `Display a QR code to connect from your mobile device.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📱 Mobile Access")
		fmt.Println("===============")
		fmt.Println("\nScan this QR code with your mobile device:")
		fmt.Println("\n╔══════════════════════════════════════╗")
		fmt.Println("║  ██████████████████████████████████  ║")
		fmt.Println("║  ██                            ██  ║")
		fmt.Println("║  ██  ▓▓▓▓  ▓▓  ▓▓▓▓  ▓▓  ▓▓▓▓  ██  ║")
		fmt.Println("║  ██  ▓▓▓▓  ▓▓  ▓▓▓▓  ▓▓  ▓▓▓▓  ██  ║")
		fmt.Println("║  ██  ▓▓  ▓▓▓▓  ▓▓  ▓▓▓▓  ▓▓  ██  ║")
		fmt.Println("║  ██  ▓▓  ▓▓▓▓  ▓▓  ▓▓▓▓  ▓▓  ██  ║")
		fmt.Println("║  ██                            ██  ║")
		fmt.Println("║  ██████████████████████████████████  ║")
		fmt.Println("╚══════════════════════════════════════╝")
		fmt.Println("\nOr visit: https://claude.ai/mobile")

		// Try to open browser
		switch runtime.GOOS {
		case "darwin":
			exec.Command("open", "https://claude.ai/mobile").Start()
		case "windows":
			exec.Command("start", "https://claude.ai/mobile").Start()
		default:
			exec.Command("xdg-open", "https://claude.ai/mobile").Start()
		}
	},
}
