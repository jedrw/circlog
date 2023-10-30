package cmd

import (
	"fmt"

	"github.com/lupinelab/circlog/config"
	"github.com/lupinelab/circlog/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "circlog [project]",
	Short: "CircleCI CLI tool",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]
		vcs, _ := cmd.Flags().GetString("vcs")
		org, _ := cmd.Flags().GetString("org")

		config, err := config.NewConfig(vcs, org)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tui.Run(config, project)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("vcs", "v", "", "Version Control System")
	rootCmd.PersistentFlags().StringP("org", "o", "", "Organisation")

	rootCmd.AddCommand(pipelinesCmd)
	rootCmd.AddCommand(workflowsCmd)
	rootCmd.AddCommand(jobsCmd)
	rootCmd.AddCommand(stepsCmd)
	rootCmd.AddCommand(logsCmd)
	cobra.EnableCommandSorting = false
}
