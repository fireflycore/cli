package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/config"
	"strings"
)

type ProtoAddModuleFormEntity struct {
	problemIndex int
	storeIndex   int

	input textinput.Model

	Module string
	Store  string
}

func (model ProtoAddModuleFormEntity) Init() tea.Cmd {
	return nil
}

func (model ProtoAddModuleFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		case "up":
			if model.problemIndex == 1 && model.storeIndex > 0 {
				model.storeIndex--
			}
		case "down":
			if model.problemIndex == 1 && (model.storeIndex < len(config.LANGUAGE)-1) {
				model.storeIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				input := model.input.Value()
				model.Module = inputReg.ReplaceAllString(input, "")
				model.problemIndex++
			case 1:
				model.Store = strings.ToLower(config.LANGUAGE[model.storeIndex])
				model.problemIndex++
			}
			if model.problemIndex+1 > len(CREATE_PROJECT_PROBLEM) {
				return model, tea.Quit
			}
		}
	}

	if model.problemIndex == 0 {
		model.input, cmd = model.input.Update(msg)
	}

	return model, cmd
}

func (model ProtoAddModuleFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(CREATE_PROJECT_PROBLEM[0])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.input.View()))
	}

	// 处理第二个问题（如果尚未回答）
	if model.problemIndex == 1 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", CREATE_PROJECT_PROBLEM[0])), PrimaryColor.Render(model.Module)))
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(CREATE_PROJECT_PROBLEM[1])))

		for ii, item := range config.LANGUAGE {
			selected := "  "
			if model.storeIndex == ii {
				selected = FocusColor.Render("->")
				item = FocusColor.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(CREATE_PROJECT_PROBLEM) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", CREATE_PROJECT_PROBLEM[0])), PrimaryColor.Render(model.Module)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", CREATE_PROJECT_PROBLEM[1])), PrimaryColor.Render(model.Store)))
	} else {
		prefix = WarningColor.Render("->")
		tips := TIPS_TEXT
		for index, item := range tips {
			str.WriteString(fmt.Sprintf("\n%s %s", prefix, item))
			if index == len(tips)-1 {
				str.WriteString("\n")
			}
		}
	}

	return str.String()
}

func NewProtoAddModule() (*ProtoAddModuleFormEntity, error) {
	input := textinput.New()
	input.Prompt = ""
	input.Focus()

	form := &ProtoAddModuleFormEntity{
		input: input,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	if form.Store == "" || form.Module == "" {
		return nil, fmt.Errorf("messing necessary params")
	}

	return form, nil
}
