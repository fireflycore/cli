package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// protoListModuleCmd represents the protoListModule command
var protoListModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "list",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protoListModule called")
	},
}

func init() {
	protoListCmd.AddCommand(protoListModuleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoListModuleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoListModuleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
