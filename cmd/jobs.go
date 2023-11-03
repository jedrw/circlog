package cmd

import (
	"fmt"

	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs [project]",
	Short: "Get the jobs for a workflow",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]
		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")
		branch, _ := cmd.Flags().GetString("branch")

		workflowId, _ := cmd.Flags().GetString("workflow-id")

		config, err := config.NewConfig(project, vcs, org, branch)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		workflowJobs, err := circleci.GetWorkflowJobs(config, workflowId)
		if err != nil {
			fmt.Println(err.Error())
		}

		outputJson(workflowJobs)
	},
}

func init() {
	jobsCmd.Flags().StringP("workflow-id", "w", "", "Workflow Id (required)")
	jobsCmd.MarkFlagRequired("workflow-id")
}
