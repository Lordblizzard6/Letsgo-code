package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagCreateCmd)
	tagCmd.AddCommand(tagDeleteCmd)
	tagCmd.AddCommand(tagPushCmd)
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage git tags",
	Long:  `Create, list, delete and push git tags.`,
}

var tagListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tags",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("git", "tag", "-l").CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(strings.TrimSpace(string(out))) == 0 {
			fmt.Println("No tags found")
			return
		}
		fmt.Println("Tags:")
		fmt.Println(string(out))
	},
}

var tagCreateCmd = &cobra.Command{
	Use:     "create <tag-name> [message]",
	Short:   "Create a new annotated tag",
	Aliases: []string{"new", "add"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: tag create <tag-name> [message]")
			os.Exit(1)
		}

		tagName := args[0]
		message := tagName
		if len(args) > 1 {
			message = strings.Join(args[1:], " ")
		}

		force, _ := cmd.Flags().GetBool("force")

		var out []byte
		var err error
		if force {
			out, err = exec.Command("git", "tag", "-fa", tagName, "-m", message).CombinedOutput()
		} else {
			out, err = exec.Command("git", "tag", "-a", tagName, "-m", message).CombinedOutput()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tag: %s\n", out)
			os.Exit(1)
		}
		fmt.Printf("✓ Created tag: %s\n", tagName)
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:     "delete <tag-name>",
	Short:   "Delete a tag",
	Aliases: []string{"del", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: tag delete <tag-name>")
			os.Exit(1)
		}

		tagName := args[0]
		out, err := exec.Command("git", "tag", "-d", tagName).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting tag: %s\n", out)
			os.Exit(1)
		}
		fmt.Printf("✓ Deleted tag: %s\n", tagName)
	},
}

var tagPushCmd = &cobra.Command{
	Use:   "push [tag-name]",
	Short: "Push tags to remote",
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")

		var out []byte
		var err error

		if all || len(args) == 0 {
			out, err = exec.Command("git", "push", "origin", "--tags").CombinedOutput()
		} else {
			tagName := args[0]
			out, err = exec.Command("git", "push", "origin", tagName).CombinedOutput()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error pushing tag: %s\n", out)
			os.Exit(1)
		}
		fmt.Println("✓ Tags pushed")
	},
}

func init() {
	tagCreateCmd.Flags().BoolP("force", "f", false, "Force replace existing tag")
	tagPushCmd.Flags().BoolP("all", "a", false, "Push all tags")
}
