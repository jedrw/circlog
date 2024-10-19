package cmd

import (
	"github.com/lupinelab/circlog/circleci"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs [project]",
	Short: "Get the jobs for a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		numPages, _ := cmd.Flags().GetInt("number-pages")
		workflowId, _ := cmd.Flags().GetString("workflow-id")
		workflowJobs, _, err := circleci.GetWorkflowJobs(cmdConfig, workflowId, numPages, "")
		if err != nil {
			return err
		}

		return outputJson(workflowJobs)
	},
}

func init() {
	jobsCmd.Flags().StringP("workflow-id", "w", "", "Workflow Id (required)")
	jobsCmd.MarkFlagRequired("workflow-id")
}
