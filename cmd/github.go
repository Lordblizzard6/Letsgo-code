package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/github"
)

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(ghLoginCmd)
	githubCmd.AddCommand(ghLogoutCmd)
	githubCmd.AddCommand(ghUserCmd)
	githubCmd.AddCommand(ghReposCmd)
	githubCmd.AddCommand(ghIssuesCmd)
	githubCmd.AddCommand(ghPullRequestsCmd)
	githubCmd.AddCommand(ghCreateIssueCmd)
	githubCmd.AddCommand(ghSearchCmd)
}

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub integration",
	Long:  `Interact with GitHub repositories, issues, and pull requests.`,
}

var ghLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using a personal access token.

To create a token:
1. Go to https://github.com/settings/tokens
2. Click "Generate new token"
3. Select scopes: repo, read:user
4. Copy the token and paste it here`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter GitHub personal access token: ")
		
		// Read token (hidden input)
		var token string
		fmt.Scanln(&token)
		token = strings.TrimSpace(token)
		
		if token == "" {
			fmt.Println("Error: Token cannot be empty")
			return
		}

		client := github.NewClient()
		if err := client.Authenticate(token); err != nil {
			fmt.Printf("Error saving token: %v\n", err)
			return
		}

		// Verify token works
		user, err := client.GetAuthenticatedUser()
		if err != nil {
			fmt.Printf("Error verifying token: %v\n", err)
			return
		}

		fmt.Printf("✓ Authenticated as %s (%s)\n", user.Login, user.HTMLURL)
	},
}

var ghLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove GitHub authentication",
	Run: func(cmd *cobra.Command, args []string) {
		client := github.NewClient()
		if err := client.Logout(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("✓ Logged out from GitHub")
	},
}

var ghUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Show authenticated user info",
	Run: func(cmd *cobra.Command, args []string) {
		client := github.NewClient()
		
		if !client.IsAuthenticated() {
			fmt.Println("Not authenticated. Run: claudego github login")
			return
		}

		user, err := client.GetAuthenticatedUser()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("User: %s\n", user.Login)
		if user.Name != "" {
			fmt.Printf("Name: %s\n", user.Name)
		}
		if user.Email != "" {
			fmt.Printf("Email: %s\n", user.Email)
		}
		if user.Bio != "" {
			fmt.Printf("Bio: %s\n", user.Bio)
		}
		fmt.Printf("Profile: %s\n", user.HTMLURL)
	},
}

var ghReposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List your repositories",
	Run: func(cmd *cobra.Command, args []string) {
		client := github.NewClient()
		
		if !client.IsAuthenticated() {
			fmt.Println("Not authenticated. Run: claudego github login")
			return
		}

		repos, err := client.ListRepos()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(repos) == 0 {
			fmt.Println("No repositories found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTARS\tFORKS\tLANGUAGE\tDESCRIPTION")
		fmt.Fprintln(w, "----\t-----\t-----\t--------\t-----------")

		for _, repo := range repos {
			desc := repo.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n", 
				repo.FullName, repo.Stars, repo.Forks, repo.Language, desc)
		}
		w.Flush()
	},
}

var ghIssuesCmd = &cobra.Command{
	Use:   "issues [owner/repo]",
	Short: "List issues for a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := args[0]
		parts := strings.Split(repoPath, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid repository format. Use: owner/repo")
			return
		}

		owner, repo := parts[0], parts[1]
		state, _ := cmd.Flags().GetString("state")

		client := github.NewClient()
		issues, err := client.ListIssues(owner, repo, state)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(issues) == 0 {
			fmt.Printf("No %s issues found\n", state)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "#\tTITLE\tSTATE\tCREATED")
		fmt.Fprintln(w, "-\t-----\t-----\t-------")

		for _, issue := range issues {
			title := issue.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", 
				issue.Number, title, issue.State, issue.CreatedAt.Format("2006-01-02"))
		}
		w.Flush()
	},
}

var ghPullRequestsCmd = &cobra.Command{
	Use:   "prs [owner/repo]",
	Short: "List pull requests for a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := args[0]
		parts := strings.Split(repoPath, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid repository format. Use: owner/repo")
			return
		}

		owner, repo := parts[0], parts[1]
		state, _ := cmd.Flags().GetString("state")

		client := github.NewClient()
		prs, err := client.ListPullRequests(owner, repo, state)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(prs) == 0 {
			fmt.Printf("No %s pull requests found\n", state)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "#\tTITLE\tAUTHOR\tBRANCH\tSTATE")
		fmt.Fprintln(w, "-\t-----\t------\t------\t-----")

		for _, pr := range prs {
			title := pr.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s -> %s\t%s\n", 
				pr.Number, title, pr.User.Login, pr.Head.Ref, pr.Base.Ref, pr.State)
		}
		w.Flush()
	},
}

var ghCreateIssueCmd = &cobra.Command{
	Use:   "create-issue [owner/repo]",
	Short: "Create a new issue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := args[0]
		parts := strings.Split(repoPath, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid repository format. Use: owner/repo")
			return
		}

		owner, repo := parts[0], parts[1]

		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")

		if title == "" {
			fmt.Println("Error: --title is required")
			return
		}

		client := github.NewClient()
		if !client.IsAuthenticated() {
			fmt.Println("Not authenticated. Run: claudego github login")
			return
		}

		issue, err := client.CreateIssue(owner, repo, title, body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ Issue created: #%d %s\n", issue.Number, issue.HTMLURL)
	},
}

var ghSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search code on GitHub",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		client := github.NewClient()
		result, err := client.SearchCode(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if result.TotalCount == 0 {
			fmt.Println("No results found")
			return
		}

		fmt.Printf("Found %d results:\n\n", result.TotalCount)

		for i, item := range result.Items {
			if i >= 10 {
				fmt.Printf("... and %d more results\n", result.TotalCount-10)
				break
			}
			fmt.Printf("%s/%s:\n", item.Repository.FullName, item.Path)
			if len(item.TextMatches) > 0 {
				fmt.Printf("  %s\n", item.TextMatches[0].Fragment)
			}
			fmt.Println()
		}
	},
}

func init() {
	ghIssuesCmd.Flags().StringP("state", "s", "open", "Issue state (open/closed/all)")
	ghPullRequestsCmd.Flags().StringP("state", "s", "open", "PR state (open/closed/all)")
	ghCreateIssueCmd.Flags().StringP("title", "t", "", "Issue title (required)")
	ghCreateIssueCmd.Flags().StringP("body", "b", "", "Issue body")
}
