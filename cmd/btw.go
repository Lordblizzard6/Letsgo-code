package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(btwCmd)
	btwCmd.Flags().BoolP("list", "l", false, "List all btw notes")
	btwCmd.Flags().BoolP("search", "s", false, "Search btw notes")
}

var btwCmd = &cobra.Command{
	Use:   "btw [note]",
	Short: "Quick note taking - By The Way",
	Long: `Take quick notes "by the way" while working with Claude.
These notes are saved and searchable for later reference.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := cmd.Flags().GetBool("list")
		search, _ := cmd.Flags().GetBool("search")

		notesDir := getBtwDir()
		os.MkdirAll(notesDir, 0755)

		if list || (len(args) == 0 && !search) {
			notes, err := listBtwNotes()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("📝 By The Way - Quick Notes")
			fmt.Println("===========================")

			if len(notes) == 0 {
				fmt.Println("No notes yet. Use 'btw <your note>' to add one.")
				return
			}

			for i, note := range notes {
				preview := note.Content
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				fmt.Printf("\n%d. %s\n", i+1, preview)
				fmt.Printf("   %s\n", note.Timestamp.Format("2006-01-02 15:04"))
			}
			return
		}

		if search {
			query := strings.Join(args, " ")
			results, err := searchBtwNotes(query)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("🔍 Search results for: %s\n\n", query)
			for i, note := range results {
				fmt.Printf("%d. %s\n", i+1, note.Content)
				fmt.Printf("   %s\n", note.Timestamp.Format("2006-01-02 15:04"))
			}
			return
		}

		// Add new note
		note := strings.Join(args, " ")
		if err := saveBtwNote(note); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving note: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Note saved!")
	},
}

type BtwNote struct {
	Content   string
	Timestamp time.Time
}

func getBtwDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "btw")
}

func saveBtwNote(content string) error {
	notesDir := getBtwDir()
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(notesDir, timestamp+".txt")

	return os.WriteFile(filename, []byte(content), 0644)
}

func listBtwNotes() ([]BtwNote, error) {
	notesDir := getBtwDir()
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return nil, err
	}

	var notes []BtwNote
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			content, _ := os.ReadFile(filepath.Join(notesDir, entry.Name()))
			timestamp, _ := time.Parse("2006-01-02_15-04-05", strings.TrimSuffix(entry.Name(), ".txt"))

			notes = append(notes, BtwNote{
				Content:   string(content),
				Timestamp: timestamp,
			})
		}
	}

	return notes, nil
}

func searchBtwNotes(query string) ([]BtwNote, error) {
	notes, err := listBtwNotes()
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var results []BtwNote

	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.Content), queryLower) {
			results = append(results, note)
		}
	}

	return results, nil
}
