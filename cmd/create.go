package cmd

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/repo"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Quick project creation.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := view.NewCreate()
		if err != nil {
			fmt.Println(err)
			return
		}

		v := store.Use.Config.Global.Version[cfg.Language]
		if v != "latest" && templateVersion == "latest" {
			templateVersion = v
		}

		rc, err := repo.New(&repo.ConfigEntity{
			Project:  cfg.Project,
			Language: cfg.Language,
			Version:  templateVersion,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		err = rc.GetRepo()
		if err != nil {
			fmt.Println(err)
			return
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

	createCmd.Flags().StringVar(&templateVersion, "version", "latest", "Template version parameter. The default value is the latest version")
}
