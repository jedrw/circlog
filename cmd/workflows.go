package cmd

import (
	"fmt"

	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows [project]",
	Short: "Get the workflows for a pipeline",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]
		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")
		branch, _ := cmd.Flags().GetString("branch")
		numPages, _ := cmd.Flags().GetInt("number-pages")

		pipelineId, _ := cmd.Flags().GetString("pipeline-id")

		config, err := config.NewConfig(project, vcs, org, branch)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		pipelineWorkflows, _, err := circleci.GetPipelineWorkflows(config, project, pipelineId, numPages, "")
		if err != nil {
			fmt.Println(err.Error())
		}

		outputJson(pipelineWorkflows)
	},
}

func init() {
	workflowsCmd.Flags().StringP("pipeline-id", "l", "", "Pipeline Id (required)")
	workflowsCmd.MarkFlagRequired("pipeline-id")
}
