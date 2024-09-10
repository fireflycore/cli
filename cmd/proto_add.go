package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/store"

	"github.com/spf13/cobra"
)

// protoAddCmd represents the protoAdd command
var protoAddCmd = &cobra.Command{
	Use: "add",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			fmt.Println(store.Use.Buf.Config.Inputs)
		} else {
			fmt.Println("current env not found buf-cli")
		}
	},
}

func init() {
	protoCmd.AddCommand(protoAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoAddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoAddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
