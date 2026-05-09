package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/tools"
)

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCmd.AddCommand(tasksStatusCmd)
	tasksCmd.AddCommand(tasksStopCmd)
	tasksCmd.AddCommand(tasksOutputCmd)
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage background tasks",
	Long:  `Create, monitor, and manage background tasks.`,
}

var tasksListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all background tasks",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		tasks := tools.ListTasks()
		if len(tasks) == 0 {
			fmt.Println("No background tasks running.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCOMMAND\tSTATUS\tSTARTED")
		fmt.Fprintln(w, "--\t-------\t------\t-------")

		for _, t := range tasks {
			status := "running"
			if t.Completed {
				if t.ExitCode == 0 {
					status = "completed"
				} else {
					status = "failed"
				}
			}
			cmd := t.Command
			if len(cmd) > 30 {
				cmd = cmd[:27] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				t.ID[:8], cmd, status, t.StartedAt.Format("15:04"))
		}
		w.Flush()
	},
}

var tasksCreateCmd = &cobra.Command{
	Use:   "create [command...]",
	Short: "Create a new background task",
	Example: `  letsgo tasks create sleep 10
  letsgo tasks create npm install
  letsgo tasks create go test ./...`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: letsgo tasks create [command...]")
			return
		}

		taskID, err := tools.CreateTask(args[0], args)
		if err != nil {
			fmt.Printf("Error creating task: %v\n", err)
			return
		}

		fmt.Printf("✓ Task created: %s\n", taskID)
		fmt.Printf("  Command: %s\n", args[0])
		fmt.Println("\nCheck status with: letsgo tasks status " + taskID)
	},
}

var tasksStatusCmd = &cobra.Command{
	Use:   "status [task-id]",
	Short: "Get status of a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		task, err := tools.GetTask(taskID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Task: %s\n", task.ID)
		fmt.Printf("Command: %s\n", task.Command)
		fmt.Printf("Started: %s\n", task.StartedAt.Format("2006-01-02 15:04:05"))

		status := "running"
		if task.Completed {
			if task.ExitCode == 0 {
				status = "completed"
			} else {
				status = "failed"
			}
			fmt.Printf("Finished: %s\n", task.FinishedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Exit Code: %d\n", task.ExitCode)
		}
		fmt.Printf("Status: %s\n", status)

		if len(task.Output) > 0 {
			fmt.Printf("\nOutput:\n%s\n", task.Output)
		}
	},
}

var tasksStopCmd = &cobra.Command{
	Use:   "stop [task-id]",
	Short: "Stop a running task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		if err := tools.StopTask(taskID); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("✓ Stopped task %s\n", taskID)
	},
}

var tasksOutputCmd = &cobra.Command{
	Use:   "output [task-id]",
	Short: "Show task output",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		task, err := tools.GetTask(taskID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(task.Output) == 0 {
			fmt.Println("No output yet.")
			return
		}

		fmt.Println(task.Output)
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsStatusCmd)
	agentsCmd.AddCommand(agentsStopCmd)
}

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage autonomous agents",
	Long:  `Monitor and control autonomous agents.`,
}

var agentsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all running agents",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		agents := tools.ListAgents()
		if len(agents) == 0 {
			fmt.Println("No agents running.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tSTARTED")
		fmt.Fprintln(w, "--\t----\t------\t-------")

		for _, a := range agents {
			status := "running"
			if a.Completed {
				status = "completed"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				a.ID[:8], a.Name, status, a.StartedAt.Format("15:04"))
		}
		w.Flush()
	},
}

var agentsStatusCmd = &cobra.Command{
	Use:   "status [agent-id]",
	Short: "Get status of an agent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentID := args[0]
		agent, err := tools.GetAgent(agentID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Agent: %s (%s)\n", agent.Name, agent.ID)
		fmt.Printf("Status: %s\n", agent.Status)
		fmt.Printf("Started: %s\n", agent.StartedAt.Format("2006-01-02 15:04:05"))

		if len(agent.Result) > 0 {
			fmt.Printf("\nResult:\n%s\n", agent.Result)
		}
	},
}

var agentsStopCmd = &cobra.Command{
	Use:   "stop [agent-id]",
	Short: "Stop a running agent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentID := args[0]
		if err := tools.StopAgent(agentID); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("✓ Stopped agent %s\n", agentID)
	},
}
