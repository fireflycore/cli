package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// protoRemoveStoreCmd represents the protoRemoveStore command
var protoRemoveStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Remove buf proto store.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protoRemoveStore called")
	},
}

func init() {
	protoRemoveCmd.AddCommand(protoRemoveStoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoRemoveStoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoRemoveStoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
