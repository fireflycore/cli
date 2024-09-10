package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"

	"github.com/spf13/cobra"
)

// protoAddStoreCmd represents the protoAddStore command
var protoAddStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Add buf proto store.",
	Run: func(cmd *cobra.Command, args []string) {
		if store.Use.Buf != nil {
			_, err := view.NewProtoAddStore()
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
	protoAddCmd.AddCommand(protoAddStoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoAddStoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoAddStoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
