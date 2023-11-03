package cmd

import (
	"fmt"

	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/spf13/cobra"
)

var stepsCmd = &cobra.Command{
	Use:   "steps [project]",
	Short: "Get the steps for a job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]
		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")
		branch, _ := cmd.Flags().GetString("branch")

		jobNumber, _ := cmd.Flags().GetInt64("job-number")

		config, err := config.NewConfig(project, vcs, org, branch)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		workflowJobs, err := circleci.GetJobSteps(config, jobNumber)
		if err != nil {
			fmt.Println(err.Error())
		}

		outputJson(workflowJobs)
	},
}

func init() {
	stepsCmd.Flags().Int64P("job-number", "j", 0, "Job Number (required)")
	stepsCmd.MarkFlagRequired("job-number")
}
