package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/cmd/internal"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create firefly microservice project",
	Long:  `quickly create a firefly microservice framework.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := internal.NewCreate()
		if err != nil {
			fmt.Println(err)
			return
		}
		internal.GetRepo()
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
