package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)


var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}