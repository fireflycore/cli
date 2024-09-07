package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runDevCmd represents the runDev command
var runDevCmd = &cobra.Command{
	Use: "dev",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("runDev called")
	},
}

func init() {
	runCmd.AddCommand(runDevCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runDevCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runDevCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
