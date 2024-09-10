package cmd

import (
	"github.com/fireflycore/cli/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

var templateVersion string

var rootCmd = &cobra.Command{
	Use:     "firefly",
	Short:   "Firefly: An elegant toolkit for Go microservices.",
	Version: config.RELEASE,
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
}
