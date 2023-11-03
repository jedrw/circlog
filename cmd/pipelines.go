package cmd

import (
	"fmt"

	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/spf13/cobra"
)

var pipelinesCmd = &cobra.Command{
	Use:   "pipelines [project]",
	Short: "Get the pipelines for a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]
		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")
		branch, _ := cmd.Flags().GetString("branch")

		config, err := config.NewConfig(project, vcs, org, branch)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		projectPipelines, err := circleci.GetProjectPipelines(config)
		if err != nil {
			fmt.Println(err.Error())
		}

		outputJson(projectPipelines)
	},
}

func init() {
	pipelinesCmd.Flags().StringP("branch", "b", "", "Branch")
}