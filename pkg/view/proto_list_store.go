package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/buf"
	"strings"
)

type ProtoListStoreFormEntity struct {
	problemIndex int

	mode      string
	modeIndex int

	module []buf.ModuleInputEntity
	local  []buf.LocalInputEntity
}

func (model *ProtoListStoreFormEntity) Init() tea.Cmd {
	return nil
}

func (model *ProtoListStoreFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				model.mode = strings.ToLower(buf.STORE_TYPE[model.modeIndex])
				model.problemIndex++
			}
			if model.problemIndex+1 > len(PROTO_LIST_STORE) {
				return model, tea.Quit
			}
		}
	}

	return model, cmd
}

func (model *ProtoListStoreFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_LIST_STORE[0])))

		for ii, item := range buf.STORE_TYPE {
			selected := "  "
			if model.modeIndex == ii {
				selected = FocusColor.Render("->")
				item = FocusColor.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(PROTO_LIST_STORE) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_LIST_STORE[0])), PrimaryColor.Render(model.mode)))
		switch model.mode {
		case "module":
			for _, item := range model.module {
				selected := "->"
				str.WriteString(fmt.Sprintf("%s %s\n", selected, item.Module))
			}
		case "local":
			for _, item := range model.local {
				selected := "->"
				str.WriteString(fmt.Sprintf("%s %s\n", selected, item.Directory))
			}
		}
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

func NewProtoListStore(module []buf.ModuleInputEntity, local []buf.LocalInputEntity) (*ProtoListStoreFormEntity, error) {
	form := &ProtoListStoreFormEntity{
		module: module,
		local:  local,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	return form, nil
}
