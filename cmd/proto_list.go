package cmd

import (
	"github.com/spf13/cobra"
)

// protoListCmd represents the protoList command
var protoListCmd = &cobra.Command{
	Use:   "list",
	Short: "store or module list",
}

func init() {
	protoCmd.AddCommand(protoListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
