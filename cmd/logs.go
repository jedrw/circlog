package cmd

import (
	"fmt"

	"github.com/jedrw/circlog/circleci"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [project]",
	Short: "Get the logs for a step",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jobNumber, _ := cmd.Flags().GetInt64("job-number")
		stepNumber, _ := cmd.Flags().GetInt64("step-number")
		stepIndex, _ := cmd.Flags().GetInt64("step-index")
		allocationId, _ := cmd.Flags().GetString("allocation-id")
		logs, err := circleci.GetStepLogs(cmdConfig, jobNumber, stepNumber, stepIndex, allocationId)
		if err != nil {
			return err
		}

		fmt.Print(logs)

		return nil
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
