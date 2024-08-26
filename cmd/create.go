package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create firefly microservice project",
	Long:  `quickly create a firefly microservice framework.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		config.Project, err = pkg.CreateProject()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		config.Language, err = pkg.CreateLanguage()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		dbs, _ := pkg.SelectDatabase(config.Language)
		if config.Database == nil {
			config.Database = []*pkg.DatabaseEntity{}
		}
		for _, item := range dbs {
			input, _ := pkg.InputDatabaseConfig(item)
			config.Database = append(config.Database, input)
		}
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
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
