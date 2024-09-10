package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runBuildCmd represents the runBuild command
var runBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "build current project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("runBuild called")
	},
}

func init() {
	runCmd.AddCommand(runBuildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runBuildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runBuildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
