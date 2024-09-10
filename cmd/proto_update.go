package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/spf13/cobra"
)

// protoUpdateCmd represents the protoUpdate command
var protoUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Generate code through buf-cli.",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			// 校验buf是否存在，存在则实行生成命令
		} else {
			fmt.Println("The buf-cli configuration is not read in the current environment")
		}
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
