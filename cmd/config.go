package cmd

import (
	"github.com/fireflycore/cli/pkg/ui"

	"github.com/spf13/cobra"
)

// configCmd represents the setConfig command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set or get global config",
	Run: func(cmd *cobra.Command, args []string) {
		_ = ui.NewConfig()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
