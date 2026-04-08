package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/tools"
)

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memorySetCmd)
	memoryCmd.AddCommand(memoryGetCmd)
	memoryCmd.AddCommand(memoryDeleteCmd)
}

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage persistent session memory",
	Long:  `Store and retrieve key-value pairs that persist across sessions.`,
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all memory entries",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		memories, err := db.ListMemory()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(memories) == 0 {
			fmt.Println("No memories stored.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tVALUE")
		fmt.Fprintln(w, "---\t-----")
		for k, v := range memories {
			if len(v) > 50 {
				v = v[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\n", k, v)
		}
		w.Flush()
	},
}

var memorySetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a memory value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, value := args[0], args[1]
		if err := db.SetMemory(key, value); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("✓ Set %s\n", key)
	},
}

var memoryGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a memory value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value, err := db.GetMemory(key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if value == "" {
			fmt.Printf("No value found for key: %s\n", key)
			return
		}
		fmt.Println(value)
	},
}

var memoryDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a memory entry",
	Aliases: []string{"rm", "del"},
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if err := db.DeleteMemory(key); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("✓ Deleted %s\n", key)
	},
}

func init() {
	rootCmd.AddCommand(costCmd)
}

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Show session cost information",
	Long:  `Display detailed cost information for the current session.`,
	Run: func(cmd *cobra.Command, args []string) {
		tracker := tools.GetCostTracker()
		sessionCost := tracker.GetSessionCost()
		
		fmt.Println("=== Cost Information ===\n")
		fmt.Printf("Current Session: %s\n", tools.FormatCost(sessionCost))
		
		// Get stats from today
		todayStats := tracker.GetUsageStats(tools.GetStartOfDay())
		fmt.Printf("Today:           %s\n", tools.FormatCost(todayStats["total_cost_usd"].(float64)))
		
		// Get stats from this week
		weekStats := tracker.GetUsageStats(tools.GetStartOfWeek())
		fmt.Printf("This Week:       %s\n", tools.FormatCost(weekStats["total_cost_usd"].(float64)))
		
		fmt.Println("\nUsage Today:")
		fmt.Printf("  Requests:  %d\n", todayStats["total_requests"])
		fmt.Printf("  Tokens:    %d (%d in, %d out)\n", 
			todayStats["total_tokens"],
			todayStats["total_input_tokens"],
			todayStats["total_output_tokens"])
	},
}
