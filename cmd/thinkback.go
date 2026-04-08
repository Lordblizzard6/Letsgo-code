package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(thinkbackCmd)
	thinkbackCmd.Flags().BoolP("list", "l", false, "List all thinkbacks")
	thinkbackCmd.Flags().BoolP("search", "s", false, "Search thinkbacks")
	thinkbackCmd.Flags().StringP("topic", "t", "", "Filter by topic")
	thinkbackCmd.Flags().IntP("limit", "n", 10, "Limit results")
}

var thinkbackCmd = &cobra.Command{
	Use:   "thinkback [query]",
	Short: "Search and recall previous conversations",
	Long: `Thinkback allows you to search through your conversation history
and recall important moments, decisions, and code snippets from past sessions.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := cmd.Flags().GetBool("list")
		searchFlag, _ := cmd.Flags().GetBool("search")
		topic, _ := cmd.Flags().GetString("topic")
		limit, _ := cmd.Flags().GetInt("limit")

		query := ""
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}

		// Get all sessions
		sessions, err := db.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
			os.Exit(1)
		}

		if list || (query == "" && !searchFlag) {
			fmt.Println("🧠 Thinkback - Recent Sessions")
			fmt.Println("================================")

			for i, s := range sessions {
				if i >= limit {
					break
				}
				history, _ := db.GetHistory(s.ID)
				fmt.Printf("\n%d. %s\n", i+1, s.Name)
				fmt.Printf("   ID: %s\n", s.ID[:8])
				fmt.Printf("   Messages: %d\n", len(history))
				fmt.Printf("   Last updated: %s\n", s.UpdatedAt.Format("2006-01-02 15:04"))
			}

			fmt.Println("\n💡 Use 'thinkback <search query>' to search your history")
			return
		}

		// Search mode
		fmt.Printf("🔍 Searching for: %s\n\n", query)

		results := searchThinkbacks(sessions, query, topic, limit)

		if len(results) == 0 {
			fmt.Println("No results found. Try a different search query.")
			return
		}

		fmt.Printf("Found %d results:\n\n", len(results))
		for i, r := range results {
			fmt.Printf("%d. [%s] %s\n", i+1, r.SessionName, r.Preview)
			fmt.Printf("   Time: %s\n", r.Timestamp.Format("2006-01-02 15:04"))
			if r.Topic != "" {
				fmt.Printf("   Topic: %s\n", r.Topic)
			}
			fmt.Println()
		}
	},
}

type ThinkbackResult struct {
	SessionName string
	Preview     string
	Timestamp   time.Time
	Topic       string
}

func searchThinkbacks(sessions []db.Session, query, topic string, limit int) []ThinkbackResult {
	var results []ThinkbackResult
	queryLower := strings.ToLower(query)

	for _, s := range sessions {
		history, _ := db.GetHistory(s.ID)

		for _, msg := range history {
			content := fmt.Sprintf("%v", msg.Content)
			contentLower := strings.ToLower(content)

			if strings.Contains(contentLower, queryLower) {
				// Extract preview
				preview := content
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}

				// Determine topic
				topicFound := ""
				if strings.Contains(contentLower, "code") || strings.Contains(contentLower, "function") {
					topicFound = "Code"
				} else if strings.Contains(contentLower, "error") || strings.Contains(contentLower, "bug") {
					topicFound = "Debugging"
				} else if strings.Contains(contentLower, "git") || strings.Contains(contentLower, "commit") {
					topicFound = "Git"
				}

				// Filter by topic if specified
				if topic != "" && !strings.EqualFold(topicFound, topic) {
					continue
				}

				results = append(results, ThinkbackResult{
					SessionName: s.Name,
					Preview:     preview,
					Timestamp:   msg.Timestamp,
					Topic:       topicFound,
				})

				if len(results) >= limit {
					return results
				}
			}
		}
	}

	return results
}
