package pkg

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
)

func SelectDatabase(lt string) ([]string, error) {
	var value []string
	prompt := &survey.MultiSelect{
		Message: "Choose databases (press space to select, a to toggle all, i/o to invert selection):",
		Options: DATABASE[lt],
	}
	err := survey.AskOne(prompt, &value)
	return value, err
}

func InputDatabaseConfig(dbType string) (*DatabaseEntity, error) {
	var value DatabaseEntity
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: fmt.Sprintf("Enter the %s database name:", dbType),
			},
			Validate: survey.Required,
		},
		{
			Name: "config",
			Prompt: &survey.Input{
				Message: fmt.Sprintf("Enter the %s database configuration file URL:", dbType),
			},
			Validate: survey.Required,
		},
	}
	err := survey.Ask(qs, &value)
	return &value, err
}
