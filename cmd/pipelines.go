package cmd

import (
	"github.com/jedrw/circlog/circleci"
	"github.com/spf13/cobra"
)

var pipelinesCmd = &cobra.Command{
	Use:   "pipelines [project]",
	Short: "Get the pipelines for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		numPages, _ := cmd.Flags().GetInt("number-pages")
		projectPipelines, _, err := circleci.GetProjectPipelines(cmdConfig, numPages, "")
		if err != nil {
			return err
		}

		return outputJson(projectPipelines)
	},
}

func init() {
	pipelinesCmd.Flags().StringP("branch", "b", "", "Branch")
}
