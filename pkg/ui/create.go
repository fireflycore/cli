package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/config"
	"github.com/fireflycore/cli/pkg/store"
	"regexp"
	"strings"
)

var inputReg = regexp.MustCompile("[^a-zA-Z0-9_]+")

type CreateFormModelEntity struct {
	Config *ConfigEntity

	problemIndex  int
	languageIndex int

	input textinput.Model
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
			if model.problemIndex == 1 && model.languageIndex > 0 {
				model.languageIndex--
			}
		case "down":
			if model.problemIndex == 1 && (model.languageIndex < len(config.LANGUAGE)-1) {
				model.languageIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				input := model.input.Value()
				model.Config.Project = inputReg.ReplaceAllString(input, "")
				model.problemIndex++
			case 1:
				model.Config.Language = strings.ToLower(config.LANGUAGE[model.languageIndex])
				model.problemIndex++
			}
			if model.problemIndex+1 > len(config.CREATE_PROBLEM) {
				return model, tea.Quit
			}
		}
	}

	if model.problemIndex == 0 {
		model.input, cmd = model.input.Update(msg)
	}

	return model, cmd
}

func (model CreateFormModelEntity) View() string {
	var str strings.Builder

	prefix := config.Color.Primary.Render("<-")
	problem := config.CREATE_PROBLEM[store.Use.Config.Global.TextLanguage]
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 && len(model.Config.Project) == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, config.Color.Info.Render(problem[0])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.input.View()))
	}

	// 处理第二个问题（如果尚未回答）
	if model.problemIndex == 1 && len(model.Config.Language) == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, config.Color.Info.Render(fmt.Sprintf("%s -", problem[0])), config.Color.Primary.Render(model.Config.Project)))
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, config.Color.Info.Render(problem[1])))

		for ii, item := range config.LANGUAGE {
			selected := "  "
			if model.languageIndex == ii {
				selected = config.Color.Focus.Render("->")
				item = config.Color.Focus.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(problem) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, config.Color.Info.Render(fmt.Sprintf("%s -", problem[0])), config.Color.Primary.Render(model.Config.Project)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, config.Color.Info.Render(fmt.Sprintf("%s -", problem[1])), config.Color.Primary.Render(model.Config.Language)))
	} else {
		prefix = config.Color.Warning.Render("->")
		tips := config.TIPS_TEXT[store.Use.Config.Global.TextLanguage]
		for index, item := range tips {
			str.WriteString(fmt.Sprintf("\n%s %s", prefix, item))
			if index == len(tips)-1 {
				str.WriteString("\n")
			}
		}
	}

	return str.String()
}

func NewCreate() (*ConfigEntity, error) {
	input := textinput.New()
	input.Prompt = ""
	input.Focus()

	form := CreateFormModelEntity{
		Config: &ConfigEntity{},

		problemIndex:  0,
		languageIndex: 0,

		input: input,
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
