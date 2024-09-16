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

	list []string
}

func (model *ProtoAddModuleFormEntity) Init() tea.Cmd {
	return nil
}

func (model *ProtoAddModuleFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		case "up":
			if model.problemIndex == 0 && model.storeIndex > 0 {
				model.storeIndex--
			}
		case "down":
			if model.problemIndex == 0 && (model.storeIndex < len(config.LANGUAGE)-1) {
				model.storeIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				model.Store = model.list[model.storeIndex]
				model.problemIndex++
			case 1:
				model.Module = model.input.Value()
				model.problemIndex++
			}
			if model.problemIndex+1 > len(PROTO_ADD_MODULE) {
				return model, tea.Quit
			}
		}
	}

	if model.problemIndex == 1 {
		model.input, cmd = model.input.Update(msg)
	}

	return model, cmd
}

func (model *ProtoAddModuleFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_ADD_MODULE[0])))

		for ii, item := range model.list {
			selected := "  "
			if model.storeIndex == ii {
				selected = FocusColor.Render("->")
				item = FocusColor.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 处理第二个问题（如果尚未回答）
	if model.problemIndex == 1 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_ADD_MODULE[0])), PrimaryColor.Render(model.Store)))
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_ADD_MODULE[1])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.input.View()))
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(PROTO_ADD_MODULE) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_ADD_MODULE[0])), PrimaryColor.Render(model.Store)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_ADD_MODULE[1])), PrimaryColor.Render(model.Module)))
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

func NewProtoAddModule(stores []string) (*ProtoAddModuleFormEntity, error) {
	input := textinput.New()
	input.Prompt = ""
	input.Focus()

	form := &ProtoAddModuleFormEntity{
		input: input,
		list:  stores,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	if form.Store == "" || form.Module == "" {
		return nil, fmt.Errorf(DangerColor.Render("messing necessary params"))
	}

	return form, nil
}
