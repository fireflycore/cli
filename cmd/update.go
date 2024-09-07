package cmd

import (
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update framework kernel",
	Long: `If executed in a directory that is not created by the firefly-cli, the global framework kernel is upgraded.
If you run the command in the directory created using the firefly-cli, the current project framework kernel will be upgraded. Please check the upgrade document before upgrading.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查当前版本是否是最新版本
		// 将api目录直接拷贝到新版本中
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
