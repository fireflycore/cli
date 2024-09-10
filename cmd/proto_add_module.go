package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// protoAddModuleCmd represents the protoAddModule command
var protoAddModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Add buf proto module",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protoAddModule called")
	},
}

func init() {
	protoAddCmd.AddCommand(protoAddModuleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoAddModuleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoAddModuleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
