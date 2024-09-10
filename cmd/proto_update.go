package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// protoUpdateCmd represents the protoUpdate command
var protoUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Generate code through buf-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protoUpdate called")
	},
}

func init() {
	protoCmd.AddCommand(protoUpdateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoUpdateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoUpdateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
