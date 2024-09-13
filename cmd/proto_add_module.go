package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"

	"github.com/spf13/cobra"
)

// protoAddModuleCmd represents the protoAddModule command
var protoAddModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Add buf proto module.",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			fmt.Println(store.Use.Buf.GetModule())
			_, err := view.NewProtoAddModule()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			fmt.Println("The buf-cli configuration is not read in the current environment")
		}
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
