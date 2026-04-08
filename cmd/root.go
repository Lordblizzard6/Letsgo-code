package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "letsgo",
	Short: "LetsGO Code - A fast Go implementation of Claude Code CLI",
	Long:  "LetsGO Code - A faster, native AI coding assistant built with Go.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.LoadConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		printHeader()
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
