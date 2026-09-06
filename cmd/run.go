package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/engine"
)

type cliDriver struct{}

func (cliDriver) Prompt(toolName string, input any) (bool, error) {
	fmt.Printf("\n[Auto-approved tool: %s]\n", toolName)
	return true, nil
}
func (cliDriver) AutoApprove(category string) bool               { return true }
func (cliDriver) SessionGranted(sessionID, category string) bool { return true }

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [prompt]",
	Short: "Run a single prompt directly without starting the interactive TUI",
	Long:  "Executes a one-shot coding task directly in your terminal, streaming the agent's actions and response.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := strings.Join(args, " ")
		eng := engine.New(cliDriver{})
		eng.Start()
		defer eng.Stop()

		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("%s %s\n\n", cyan("⚡ Task:"), prompt)
		eng.Send(engine.SendMessage{Text: prompt, SessionID: eng.SessionID()})

		for ev := range eng.Events() {
			switch e := ev.(type) {
			case engine.StreamDelta:
				fmt.Print(e.Text)
			case engine.ToolExecuting:
				fmt.Printf("\n%s Executing tool %s...\n", yellow("▶"), e.ToolCallID)
			case engine.ToolResult:
				if e.IsError {
					fmt.Printf("%s Error: %s\n", yellow("✖"), e.Content)
				} else {
					fmt.Printf("%s Done.\n", yellow("✔"))
				}
			case engine.ErrorEvent:
				fmt.Fprintf(os.Stderr, "\nError: %s\n", e.Message)
				return
			case engine.Idle:
				fmt.Println()
				return
			}
		}
	},
}
