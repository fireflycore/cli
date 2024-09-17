package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// protoRemoveModuleCmd represents the protoRemoveModule command
var protoRemoveModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Remove buf proto module.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protoRemoveModule called")
	},
}

func init() {
	protoRemoveCmd.AddCommand(protoRemoveModuleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoRemoveModuleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoRemoveModuleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
