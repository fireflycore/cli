package internal

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"strings"
)

var config ConfigEntity

type FormModelEntity struct {
	askIndex      int
	languageIndex int
	databaseFocus int

	projectInput textinput.Model
}

func (model FormModelEntity) Init() tea.Cmd {
	return nil
}

func (model FormModelEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if model.askIndex == 1 && (model.languageIndex < len(LANGUAGE)-1) {
				model.languageIndex++
			}
			if model.askIndex == 2 && (model.databaseFocus < len(DatabaseList)-1) {
				model.databaseFocus++
			}
		case "enter":
			switch model.askIndex {
			case 0:
				config.Project = model.projectInput.View()
				model.askIndex++
			case 1:
				config.Language = LANGUAGE[model.languageIndex]
				if list, ok := Database[config.Language]; ok {
					DatabaseList = list
				}
				model.askIndex++
			case 2:
				model.askIndex++
			}
			if model.askIndex+1 > len(ASK) {
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

func (model FormModelEntity) View() string {
	var str strings.Builder

	// 处理第一个问题（如果尚未回答）
	if model.askIndex == 0 && len(config.Project) == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", primary.Render("<-"), info.Render(ASK[0])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.projectInput.View()))
	}

	// 处理第二个问题（如果尚未回答）
	if model.askIndex == 1 && len(config.Language) == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", ASK[0])), primary.Render(config.Project)))
		str.WriteString(fmt.Sprintf("%s %s\n", primary.Render("<-"), info.Render(ASK[1])))

		for ii, item := range LANGUAGE {
			selected := "  "
			if model.languageIndex == ii {
				selected = focus.Render("->")
				item = focus.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 处理第三个问题（如果尚未回答）
	if model.askIndex == 2 && len(config.Database) == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", ASK[0])), primary.Render(config.Project)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", ASK[1])), primary.Render(config.Language)))
		str.WriteString(fmt.Sprintf("%s %s\n", primary.Render("<-"), info.Render(ASK[2])))

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
	if model.askIndex == len(ASK) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", ASK[0])), primary.Render(config.Project)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", ASK[1])), primary.Render(config.Language)))

		var dbs []string
		for _, item := range DatabaseList {
			dbs = append(dbs, item.Type)
		}
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", ASK[2])), primary.Render(strings.Join(dbs, ", "))))
	}

	prefix := warning.Render("->")
	str.WriteString(fmt.Sprintf("\n%s ctrl+c or q to exit the cli.\n%s enter confirm or next step.\n", prefix, prefix))

	return str.String()
}

func New() {
	projectInput := textinput.New()
	projectInput.Prompt = ""
	projectInput.Focus()

	form := FormModelEntity{
		askIndex:      0,
		languageIndex: 0,
		databaseFocus: 0,

		projectInput: projectInput,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(config.Project, config.Language)
}
