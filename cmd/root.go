package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg"
	"os"

	"github.com/spf13/cobra"
)

var config pkg.ConfigEntity

var rootCmd = &cobra.Command{
	Use:   "firefly",
	Short: "firefly cli",
	Long:  `firefly cli tools`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("firefly-cli run hooks")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	//rootCmd.PersistentFlags().StringVar(&project, "project", "", "set project name")
	//rootCmd.PersistentFlags().StringVar(&language, "language", "", "set develop language")
}
