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
	rootCmd.AddCommand(thinkbackPlayCmd)
	thinkbackPlayCmd.Flags().StringP("session", "s", "", "Session ID to play")
	thinkbackPlayCmd.Flags().IntP("speed", "", 1, "Playback speed (1-3)")
}

var thinkbackPlayCmd = &cobra.Command{
	Use:   "thinkback-play",
	Short: "Replay a previous conversation",
	Long: `Replay a previous conversation session to review the flow
of ideas, decisions, and code changes. Useful for reviewing past work.`,
	Run: func(cmd *cobra.Command, args []string) {
		sessionID, _ := cmd.Flags().GetString("session")
		speed, _ := cmd.Flags().GetInt("speed")

		if sessionID == "" {
			// List available sessions
			sessions, _ := listThinkbackSessions()
			if len(sessions) == 0 {
				fmt.Println("No sessions available for playback.")
				return
			}

			fmt.Println("📼 Thinkback Play - Available Sessions")
			fmt.Println("========================================")

			for i, s := range sessions {
				fmt.Printf("\n%d. %s\n", i+1, s.Name)
				fmt.Printf("   ID: %s\n", s.ID)
				fmt.Printf("   Run: thinkback-play --session %s\n", s.ID)
			}
			return
		}

		// Playback mode
		fmt.Printf("▶️  Playing back session: %s\n", sessionID[:8])
		fmt.Printf("   Speed: %dx\n\n", speed)

		// Load session history
		history, err := loadSessionHistory(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading session: %v\n", err)
			os.Exit(1)
		}

		if len(history) == 0 {
			fmt.Println("No messages in this session.")
			return
		}

		fmt.Printf("📊 Session contains %d messages\n\n", len(history))

		// Playback loop
		for i, msg := range history {
			role := "👤 You"
			if msg.Role == "assistant" {
				role = "🤖 Claude"
			}

			content := fmt.Sprintf("%v", msg.Content)

			fmt.Printf("[%d/%d] %s:\n", i+1, len(history), role)
			fmt.Println(strings.Repeat("-", 50))

			// Print with word wrap
			printWrapped(content, 50)

			fmt.Println()

			// Delay based on speed
			if speed > 0 && i < len(history)-1 {
				time.Sleep(time.Duration(1000/speed) * time.Millisecond)
			}
		}

		fmt.Println("✅ Playback complete")
	},
}

type ThinkbackSession struct {
	ID   string
	Name string
}

func listThinkbackSessions() ([]ThinkbackSession, error) {
	// Get from db
	sessions, err := db.ListSessions()
	if err != nil {
		return nil, err
	}

	var result []ThinkbackSession
	for _, s := range sessions {
		result = append(result, ThinkbackSession{
			ID:   s.ID,
			Name: s.Name,
		})
	}
	return result, nil
}

type HistoryMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

func loadSessionHistory(sessionID string) ([]HistoryMessage, error) {
	// Load from db
	dbHistory, err := db.GetHistory(sessionID)
	if err != nil {
		return nil, err
	}

	var result []HistoryMessage
	for _, msg := range dbHistory {
		content := fmt.Sprintf("%v", msg.Content)
		result = append(result, HistoryMessage{
			Role:      msg.Role,
			Content:   content,
			Timestamp: msg.Timestamp,
		})
	}
	return result, nil
}

func printWrapped(text string, width int) {
	words := strings.Fields(text)
	line := ""

	for _, word := range words {
		if len(line)+len(word)+1 > width {
			fmt.Println(line)
			line = word
		} else {
			if line != "" {
				line += " "
			}
			line += word
		}
	}

	if line != "" {
		fmt.Println(line)
	}
}
