package pkg

import (
	"github.com/AlecAivazis/survey/v2"
)

func CreateProject() (string, error) {
	value := ""
	prompt := &survey.Input{Message: "What is your project name?"}
	err := survey.AskOne(prompt, &value, survey.WithValidator(survey.Required))
	return value, err
}

func CreateLanguage() (string, error) {
	value := ""
	prompt := &survey.Select{
		Message: "Choose your develop language:",
		Options: LANGUAGE,
		Default: LANGUAGE[0],
	}
	err := survey.AskOne(prompt, &value)
	return value, err
}
