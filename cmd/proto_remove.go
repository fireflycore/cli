package cmd

import (
	"github.com/spf13/cobra"
)

// protoRemoveCmd represents the protoRemove command
var protoRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove proto store or module.",
}

func init() {
	protoCmd.AddCommand(protoRemoveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoRemoveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoRemoveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
