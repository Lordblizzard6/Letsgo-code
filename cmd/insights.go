package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(insightsCmd)
	insightsCmd.Flags().BoolP("sessions", "s", false, "Show session insights")
	insightsCmd.Flags().BoolP("models", "m", false, "Show model usage breakdown")
	insightsCmd.Flags().StringP("since", "", "7d", "Time range (1d, 7d, 30d)")
}

var insightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Generate usage insights and analytics",
	Long:  `Analyze your Claude Code usage patterns, costs, and productivity metrics.`,
	Run: func(cmd *cobra.Command, args []string) {
		sessions, _ := cmd.Flags().GetBool("sessions")
		models, _ := cmd.Flags().GetBool("models")
		since, _ := cmd.Flags().GetString("since")

		fmt.Println("📊 Claude Code Insights")
		fmt.Println("======================")
		fmt.Printf("Period: last %s\n\n", since)

		// Load all sessions
		sessionList, err := db.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
			os.Exit(1)
		}

		if len(sessionList) == 0 {
			fmt.Println("No sessions found. Start using Claude Code to generate insights!")
			return
		}

		// Calculate metrics
		totalMessages := 0
		sessionDurations := make(map[string]time.Duration)
		modelUsage := make(map[string]int)

		for _, s := range sessionList {
			history, _ := db.GetHistory(s.ID)
			totalMessages += len(history)
			sessionDurations[s.ID] = time.Since(s.CreatedAt)
		}

		// Overall stats
		fmt.Println("📈 Overall Statistics")
		fmt.Printf("  Total Sessions: %d\n", len(sessionList))
		fmt.Printf("  Total Messages: %d\n", totalMessages)
		if len(sessionList) > 0 {
			fmt.Printf("  Avg Messages/Session: %.1f\n", float64(totalMessages)/float64(len(sessionList)))
		}
		fmt.Println()

		// Session breakdown
		if sessions || (!sessions && !models) {
			fmt.Println("📝 Session Breakdown")
			sort.Slice(sessionList, func(i, j int) bool {
				return sessionList[i].UpdatedAt.After(sessionList[j].UpdatedAt)
			})

			for i, s := range sessionList {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(sessionList)-5)
					break
				}
				history, _ := db.GetHistory(s.ID)
				duration := sessionDurations[s.ID]
				fmt.Printf("  • %s - %d msgs - %s\n", s.Name, len(history), formatDuration(duration))
			}
			fmt.Println()
		}

		// Activity by time
		if !models {
			fmt.Println("⏰ Activity Patterns")
			now := time.Now()
			todayCount := 0
			weekCount := 0

			for _, s := range sessionList {
				if s.UpdatedAt.Day() == now.Day() && s.UpdatedAt.Month() == now.Month() {
					todayCount++
				}
				if now.Sub(s.UpdatedAt) < 7*24*time.Hour {
					weekCount++
				}
			}

			fmt.Printf("  Sessions today: %d\n", todayCount)
			fmt.Printf("  Sessions this week: %d\n", weekCount)
			fmt.Println()
		}

		// Model usage (placeholder)
		if models {
			fmt.Println("🤖 Model Usage")
			fmt.Println("  (Model tracking would appear here)")
			_ = modelUsage
			fmt.Println()
		}

		// Recommendations
		fmt.Println("💡 Recommendations")
		if len(sessionList) > 10 {
			fmt.Println("  • Consider archiving old sessions to improve performance")
		}
		if totalMessages > 1000 {
			fmt.Println("  • You've been very productive! Consider reviewing past sessions")
		}
		fmt.Println("  • Use 'compact' regularly to keep context manageable")
	},
}

func formatDurationInsights(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
