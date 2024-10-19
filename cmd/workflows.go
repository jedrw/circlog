package cmd

import (
	"github.com/lupinelab/circlog/circleci"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows [project]",
	Short: "Get the workflows for a pipeline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		numPages, _ := cmd.Flags().GetInt("number-pages")
		pipelineId, _ := cmd.Flags().GetString("pipeline-id")
		pipelineWorkflows, _, err := circleci.GetPipelineWorkflows(cmdConfig, pipelineId, numPages, "")
		if err != nil {
			return err
		}

		return outputJson(pipelineWorkflows)
	},
}

func init() {
	workflowsCmd.Flags().StringP("pipeline-id", "l", "", "Pipeline Id (required)")
	workflowsCmd.MarkFlagRequired("pipeline-id")
}
