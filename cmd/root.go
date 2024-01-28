package cmd

import (
	"github.com/lupinelab/circlog/config"
	"github.com/lupinelab/circlog/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "circlog [project]",
	Short: "CircleCI CLI tool",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var project string

		if len(args) > 0 {
			project = args[0]
		} else {
			project = ""
		}

		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")
		branch, _ := cmd.Flags().GetString("branch")

		config, err := config.NewConfig(project, vcs, org, branch)
		if err != nil {
			return err
		}

		circlogTui := tui.NewCirclogTui(config)

		return circlogTui.Run()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("vcs", "v", "", "Version Control System")
	rootCmd.PersistentFlags().StringP("org", "o", "", "Organisation")
	rootCmd.PersistentFlags().IntP("number-pages", "n", 1, "Number of pages to return. -1 to return everything, this may take a long time if the project has many pipelines")
	rootCmd.Flags().StringP("branch", "b", "", "Branch")

	rootCmd.AddCommand(pipelinesCmd)
	rootCmd.AddCommand(workflowsCmd)
	rootCmd.AddCommand(jobsCmd)
	rootCmd.AddCommand(stepsCmd)
	rootCmd.AddCommand(logsCmd)
	cobra.EnableCommandSorting = false
}
