package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/config"
	"regexp"
	"strings"
)

var inputReg = regexp.MustCompile("[^a-zA-Z0-9_]+")

type CreateFormModelEntity struct {
	Config *ConfigEntity

	askIndex      int
	languageIndex int
	databaseFocus int

	projectInput textinput.Model
}

func (model CreateFormModelEntity) Init() tea.Cmd {
	return nil
}

func (model CreateFormModelEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		case "up":
			if model.askIndex == 1 && model.languageIndex > 0 {
				model.languageIndex--
			}
			if model.askIndex == 2 && model.databaseFocus > 0 {
				model.databaseFocus--
			}
		case "down":
			if model.askIndex == 1 && (model.languageIndex < len(config.LANGUAGE)-1) {
				model.languageIndex++
			}
			if model.askIndex == 2 && (model.databaseFocus < len(DatabaseList)-1) {
				model.databaseFocus++
			}
		case "enter":
			switch model.askIndex {
			case 0:
				input := model.projectInput.Value()
				model.Config.Project = inputReg.ReplaceAllString(input, "")
				model.askIndex++
			case 1:
				model.Config.Language = strings.ToLower(config.LANGUAGE[model.languageIndex])
				if list, ok := Database[model.Config.Language]; ok {
					DatabaseList = list
				}
				model.askIndex++
			case 2:
				model.askIndex++
			}
			if model.askIndex+1 > len(CREATE_PROBLEM) {
				return model, tea.Quit
			}
		case " ":
			if model.askIndex == 2 {
				DatabaseList[model.databaseFocus].Select = !DatabaseList[model.databaseFocus].Select
			}
		}
	}

	if model.askIndex == 0 {
		model.projectInput, cmd = model.projectInput.Update(msg)
	}

	return model, cmd
}

func (model CreateFormModelEntity) View() string {
	var str strings.Builder

	// 处理第一个问题（如果尚未回答）
	if model.askIndex == 0 && len(model.Config.Project) == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", primary.Render("<-"), info.Render(CREATE_PROBLEM[0])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.projectInput.View()))
	}

	// 处理第二个问题（如果尚未回答）
	if model.askIndex == 1 && len(config.LANGUAGE) == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", CREATE_PROBLEM[0])), primary.Render(model.Config.Project)))
		str.WriteString(fmt.Sprintf("%s %s\n", primary.Render("<-"), info.Render(CREATE_PROBLEM[1])))

		for ii, item := range config.LANGUAGE {
			selected := "  "
			if model.languageIndex == ii {
				selected = focus.Render("->")
				item = focus.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 处理第三个问题（如果尚未回答）
	if model.askIndex == 2 && len(model.Config.Database) == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", CREATE_PROBLEM[0])), primary.Render(model.Config.Project)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", CREATE_PROBLEM[1])), primary.Render(model.Config.Language)))
		str.WriteString(fmt.Sprintf("%s %s\n", primary.Render("<-"), info.Render(CREATE_PROBLEM[2])))

		for ii, item := range DatabaseList {
			selected := "[ ]"
			value := item.Type

			if model.databaseFocus == ii && item.Select {
				selected = info.Render("[*]")
				value = info.Render(value)
			} else if model.databaseFocus == ii {
				selected = info.Render(selected)
				value = info.Render(value)
			} else if item.Select {
				selected = focus.Render("[*]")
				value = focus.Render(value)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, value))
		}

		if len(DatabaseList) == 0 {
			str.WriteString("Database configuration is not implemented under this development language.\n")
		}
	}

	// 如果已经完成了所有问题
	if model.askIndex == len(CREATE_PROBLEM) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", CREATE_PROBLEM[0])), primary.Render(model.Config.Project)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", CREATE_PROBLEM[1])), primary.Render(model.Config.Language)))

		var dbs []string
		for _, item := range DatabaseList {
			dbs = append(dbs, item.Type)
		}
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", CREATE_PROBLEM[2])), primary.Render(strings.Join(dbs, ", "))))
	}

	prefix := warning.Render("->")
	str.WriteString(fmt.Sprintf("\n%s ctrl+c or q to exit the cli.\n%s enter confirm or next step.\n", prefix, prefix))

	return str.String()
}

func NewCreate() (*ConfigEntity, error) {
	projectInput := textinput.New()
	projectInput.Prompt = ""
	projectInput.Focus()

	form := CreateFormModelEntity{
		Config: &ConfigEntity{},

		askIndex:      0,
		languageIndex: 0,
		databaseFocus: 0,

		projectInput: projectInput,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	if form.Config.Project == "" || form.Config.Language == "" {
		return nil, fmt.Errorf("messing necessary params")
	}

	return form.Config, nil
}
