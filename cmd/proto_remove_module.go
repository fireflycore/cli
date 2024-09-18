package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"

	"github.com/spf13/cobra"
)

// protoRemoveModuleCmd represents the protoRemoveModule command
var protoRemoveModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Remove buf proto module",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			form, err := view.NewProtoRemoveModule(store.Use.Buf.Config.GetModuleStores())
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			store.Use.Buf.Config.RemoveGenModule(form.Store, form.Module)
			_ = store.Use.Buf.WriteConfig()
		} else {
			fmt.Println("The buf-cli configuration is not read in the current environment")
		}
	},
}

func init() {
	protoRemoveCmd.AddCommand(protoRemoveModuleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoRemoveModuleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoRemoveModuleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
