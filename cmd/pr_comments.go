package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(prCommentsCmd)
	prCommentsCmd.Flags().String("pr", "", "PR number or URL")
}

var prCommentsCmd = &cobra.Command{
	Use:   "pr-comments",
	Short: "View and manage PR comments",
	Long:  `View, reply to, and manage pull request comments.`,
	Run: func(cmd *cobra.Command, args []string) {
		pr, _ := cmd.Flags().GetString("pr")
		if pr == "" {
			fmt.Println("Usage: pr-comments --pr <number>")
			os.Exit(1)
		}

		fmt.Printf("📋 PR Comments for PR #%s\n", pr)
		fmt.Println("========================")
		fmt.Println()
		fmt.Println("No comments to display.")
		fmt.Println("\n(Integrate with GitHub API to fetch real comments)")
	},
}
