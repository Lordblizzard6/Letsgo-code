package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/analytics"
	"github.com/user/go-claude-code/internal/updater"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statsCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		checker := updater.NewUpdateChecker()
		release, err := checker.CheckForUpdates()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		if release == nil {
			fmt.Printf("✓ You are running the latest version (%s)\n", updater.GetCurrentVersion())
			return
		}

		fmt.Printf("\n🚀 New version available: %s\n", release.Version)
		fmt.Printf("Current version: %s\n\n", updater.GetCurrentVersion())
		fmt.Printf("Release: %s\n", release.Name)
		fmt.Printf("Published: %s\n\n", release.PublishedAt.Format("2006-01-02"))

		if release.Body != "" {
			fmt.Println("Changelog:")
			fmt.Println(release.Body[:min(len(release.Body), 500)])
			if len(release.Body) > 500 {
				fmt.Println("...")
			}
		}

		fmt.Printf("\nDownload: %s\n", release.HTMLURL)
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Run: func(cmd *cobra.Command, args []string) {
		a := analytics.GetAnalytics()
		stats := a.GetStats()

		fmt.Println("=== Usage Statistics ===")

		// Overall stats
		fmt.Printf("Total Sessions:     %d\n", stats.TotalSessions)
		fmt.Printf("Total Duration:       %s\n", formatDuration(stats.TotalDuration))
		fmt.Printf("Total Messages:       %d\n", stats.TotalMessages)
		fmt.Printf("Total Tokens:         %d\n", stats.TotalTokens)
		fmt.Printf("Total Cost:           $%.4f\n", stats.TotalCost)
		fmt.Printf("Last Updated:         %s\n\n", stats.LastUpdated.Format("2006-01-02 15:04"))

		// Current session
		current := a.GetCurrentSession()
		if current != nil {
			fmt.Println("=== Current Session ===")
			fmt.Printf("Started:    %s\n", current.StartTime.Format("15:04:05"))
			fmt.Printf("Duration:   %s\n", formatDuration(time.Since(current.StartTime)))
			fmt.Printf("Messages:   %d\n", current.MessagesSent)
			fmt.Printf("Tokens:     %d\n", current.TokensUsed)
			fmt.Printf("Cost:       $%.4f\n\n", current.Cost)
		}

		// Top tools
		topTools := a.GetTopTools(10)
		if len(topTools) > 0 {
			fmt.Println("=== Top Tools Used ===")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, tool := range topTools {
				fmt.Fprintf(w, "  %s\t%d uses\n", tool.Name, tool.Count)
			}
			w.Flush()
			fmt.Println()
		}

		// Daily stats (last 7 days)
		dailyStats := a.GetDailyStats()
		if len(dailyStats) > 0 {
			fmt.Println("=== Daily Activity ===")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "DATE\tSESSIONS\tTOKENS\tCOST")

			// Get last 7 days
			for i := 6; i >= 0; i-- {
				date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
				stat, exists := dailyStats[date]
				if !exists {
					stat = analytics.DailyStat{Date: date}
				}
				fmt.Fprintf(w, "%s\t%d\t%d\t$%.2f\n",
					stat.Date, stat.Sessions, stat.Tokens, stat.Cost)
			}
			w.Flush()
		}
	},
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
