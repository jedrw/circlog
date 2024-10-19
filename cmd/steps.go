package cmd

import (
	"github.com/lupinelab/circlog/circleci"
	"github.com/spf13/cobra"
)

var stepsCmd = &cobra.Command{
	Use:   "steps [project]",
	Short: "Get the steps for a job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jobNumber, _ := cmd.Flags().GetInt64("job-number")
		workflowJobs, err := circleci.GetJobSteps(cmdConfig, jobNumber)
		if err != nil {
			return err
		}

		return outputJson(workflowJobs)
	},
}

func init() {
	stepsCmd.Flags().Int64P("job-number", "j", 0, "Job Number (required)")
	stepsCmd.MarkFlagRequired("job-number")
}
