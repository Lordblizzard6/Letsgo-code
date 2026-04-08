package cmd

import (
	"fmt"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/tui"
)

func init() {
	rootCmd.AddCommand(chatCmd)
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with Let's Go Code",
	Long:  "Start an interactive TUI chat session with the AI assistant",
	Run: func(cmd *cobra.Command, args []string) {
		printChatHeader()
		if err := tui.RunChat(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting chat: %v\n", err)
			os.Exit(1)
		}
	},
}

// printChatHeader muestra el ASCII art de bienvenida para el chat
func printChatHeader() {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Println()
	// Generar ASCII art con figlet
	myFigure := figure.NewFigure("Lets Go", "slant", true)
	fmt.Println(green(myFigure.String()))

	myFigure2 := figure.NewFigure("Code", "slant", true)
	fmt.Println(cyan(myFigure2.String()))

	fmt.Println()
	fmt.Println(white("        A faster, native AI coding assistant built with Go"))
	fmt.Println()
	fmt.Println(green("              Starting interactive chat session..."))
	fmt.Println(white("              Type /help for available commands, Ctrl+C to exit"))
	fmt.Println()
}
