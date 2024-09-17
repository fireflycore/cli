package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"

	"github.com/spf13/cobra"
)

// protoListStoreCmd represents the protoListStore command
var protoListStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "list",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			_, err := view.NewProtoListStore(store.Use.Buf.GetModuleStores(), store.Use.Buf.GetLocalStores())
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
	protoListCmd.AddCommand(protoListStoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoListStoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoListStoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
