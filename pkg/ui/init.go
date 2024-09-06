package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/config"
	"strings"
)

type InitFormModelEntity struct {
	problemIndex      int
	textLanguageIndex int

	textLanguage string
}

func (model InitFormModelEntity) Init() tea.Cmd {
	return nil
}

func (model InitFormModelEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		case "up":
			if model.problemIndex == 0 && model.textLanguageIndex > 0 {
				model.textLanguageIndex--
			}
		case "down":
			if model.problemIndex == 0 && (model.textLanguageIndex < len(config.TEXT_LANGUAGE)-1) {
				model.textLanguageIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				model.textLanguage = strings.ToLower(config.TEXT_LANGUAGE[model.textLanguageIndex])
				model.problemIndex++
			}
			if model.problemIndex+1 > len(config.InitProblemTextLang) {
				return model, tea.Quit
			}
		}
	}

	return model, cmd
}

func (model InitFormModelEntity) View() string {
	var str strings.Builder

	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 && len(model.textLanguage) == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", config.InitProblemTextLang[0])), primary.Render(model.textLanguage)))

		for ii, item := range config.LANGUAGE {
			selected := "  "
			if model.textLanguageIndex == ii {
				selected = focus.Render("->")
				item = focus.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(CREATE_PROBLEM) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", primary.Render("<-"), info.Render(fmt.Sprintf("%s -", config.InitProblemTextLang[0])), primary.Render(model.textLanguage)))

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

func NewInit() (*ConfigEntity, error) {
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
