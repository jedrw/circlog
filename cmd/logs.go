package cmd

import (
	"fmt"

	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [project]",
	Short: "Get the logs for a step",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]
		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")
		branch, _ := cmd.Flags().GetString("branch")

		jobNumber, _ := cmd.Flags().GetInt64("job-number")
		stepNumber, _ := cmd.Flags().GetInt64("step-number")
		stepIndex, _ := cmd.Flags().GetInt64("step-index")
		allocationId, _ := cmd.Flags().GetString("allocation-id")

		config, err := config.NewConfig(project, vcs, org, branch)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		logs, err := circleci.GetStepLogs(config, jobNumber, stepNumber, stepIndex, allocationId)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Print(logs)
	},
}

func init() {
	logsCmd.Flags().Int64P("job-number", "j", 0, "Job Number (required)")
	logsCmd.Flags().Int64P("step-number", "s", 0, "Step Number (required)")
	logsCmd.Flags().Int64P("step-index", "i", 0, "Step Index (required)")
	logsCmd.Flags().StringP("allocation-id", "a", "", "Allocation Id (required)")

	logsCmd.MarkFlagRequired("job-number")
	logsCmd.MarkFlagRequired("step-number")
	logsCmd.MarkFlagRequired("step-index")
	logsCmd.MarkFlagRequired("allocation-id")
}
