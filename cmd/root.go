package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "letsgo [prompt]",
	Short: "LetsGO Code - An open, fast AI coding assistant built with Go",
	Long:  "LetsGO Code - An AI coding assistant harness built with Go, TUI (BubbleTea), and GUI (Wails).",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.LoadConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			runCmd.Run(cmd, args)
			return
		}
		printChatHeader()
		if err := tui.RunChat(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting chat: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root flags if needed
}

// printHeader muestra el ASCII art de LetsGO Code con colores
func printHeader() {
	// Definir colores
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	purple := color.New(color.FgMagenta, color.Bold).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Println()
	// ASCII art coloreado línea por línea
	// "Let's" en amarillo, "Go" en cyan, "Code" en morado
	fmt.Println(yellow(`    __`) + cyan(`         ______`) + purple(`       ____`))
	fmt.Println(yellow(`   / /   ___`) + cyan(`  / ____/____`) + purple(`  / __ \`) + white(`_________  ________  __________`))
	fmt.Println(yellow(`  / /   / _ \`) + cyan(`/ / __/ ___/`) + purple(` / / / /`) + white(` ___/ __ \/ ___/ _ \/ ___/ ___/`))
	fmt.Println(yellow(` / /___/  __/`) + cyan(` /_/ / /`) + purple(`    / /_/ /`) + white(` /  / /_/ / /  /  __(__  |__  )`))
	fmt.Println(yellow(`/_____/\___/`) + cyan(`\____/_/`) + purple(`    /_____/_/`) + white(`   \____/_/   \___/____/____/`))
	fmt.Println()
	fmt.Println(white("A faster, native AI coding assistant built with Go"))
	fmt.Println()
	fmt.Println("Usage: letsgo <command> or letsgo chat to start interactive mode")
	fmt.Println()
}
