package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/buf"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"

	"github.com/spf13/cobra"
)

// protoRemoveStoreCmd represents the protoRemoveStore command
var protoRemoveStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Remove buf proto store",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			form, err := view.NewProtoRemoveStore(buf.STORE_TYPE, store.Use.Buf.Config.GetModuleStores(), store.Use.Buf.Config.GetLocalStores())
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			store.Use.Buf.Config.RemoveGenStore(form.Mode, form.Store)
			_ = store.Use.Buf.WriteConfig()
		} else {
			fmt.Println("The buf-cli configuration is not read in the current environment")
		}
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
