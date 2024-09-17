package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// protoListStoreCmd represents the protoListStore command
var protoListStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "list",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protoListStore called")
	},
}

func init() {
	protoListCmd.AddCommand(protoListStoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoListStoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoListStoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
