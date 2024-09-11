package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/buf"
	"strings"
)

type ProtoAddStoreFormEntity struct {
	problemIndex int
	modeIndex    int

	input textinput.Model

	Mode  string
	Store string
}

func (model *ProtoAddStoreFormEntity) Init() tea.Cmd {
	return nil
}

func (model *ProtoAddStoreFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		case "up":
			if model.problemIndex == 0 && model.modeIndex > 0 {
				model.modeIndex--
			}
		case "down":
			if model.problemIndex == 0 && (model.modeIndex < len(buf.STORE_TYPE)-1) {
				model.modeIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				model.Mode = strings.ToLower(buf.STORE_TYPE[model.modeIndex])
				model.problemIndex++
			case 1:
				input := model.input.Value()
				model.Store = inputReg.ReplaceAllString(input, "")
				model.problemIndex++
			}
			if model.problemIndex+1 > len(PROTO_ADD_STORE) {
				return model, tea.Quit
			}
		}
	}

	if model.problemIndex == 1 {
		model.input, cmd = model.input.Update(msg)
	}

	return model, cmd
}

func (model *ProtoAddStoreFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_ADD_STORE[0])))

		for ii, item := range buf.STORE_TYPE {
			selected := "  "
			if model.modeIndex == ii {
				selected = FocusColor.Render("->")
				item = FocusColor.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 处理第二个问题（如果尚未回答）
	if model.problemIndex == 1 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_ADD_STORE[0])), PrimaryColor.Render(model.Mode)))
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_ADD_STORE[1])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.input.View()))
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(PROTO_ADD_STORE) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_ADD_STORE[0])), PrimaryColor.Render(model.Mode)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_ADD_STORE[1])), PrimaryColor.Render(model.Store)))
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

func NewProtoAddStore() (*ProtoAddStoreFormEntity, error) {
	input := textinput.New()
	input.Prompt = ""
	input.Focus()

	form := &ProtoAddStoreFormEntity{
		input: input,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	if form.Mode == "" || form.Store == "" {
		return nil, fmt.Errorf(DangerColor.Render("messing necessary params"))
	}

	return form, nil
}
