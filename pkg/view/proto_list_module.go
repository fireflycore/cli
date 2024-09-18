package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/buf"
	"strings"
)

type ProtoListModuleFormEntity struct {
	problemIndex int

	store      string
	storeIndex int

	stores []buf.ModuleInputEntity
	types  []string
}

func (model *ProtoListModuleFormEntity) Init() tea.Cmd {
	return nil
}

func (model *ProtoListModuleFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if model.problemIndex == 0 && (model.storeIndex < len(model.stores)-1) {
				model.storeIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				store := model.stores[model.storeIndex]
				model.store = store.Module
				model.types = store.Types
				model.problemIndex++
			}
			if model.problemIndex+1 > len(PROTO_LIST_MODULE) {
				return model, tea.Quit
			}
		}
	}

	return model, cmd
}

func (model *ProtoListModuleFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_LIST_MODULE[0])))

		for ii, item := range model.stores {
			selected := "  "
			if model.storeIndex == ii {
				selected = FocusColor.Render("->")
				item.Module = FocusColor.Render(item.Module)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item.Module))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(PROTO_LIST_MODULE) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_LIST_MODULE[0])), PrimaryColor.Render(model.store)))
		for _, item := range model.types {
			selected := "->"
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
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

func NewProtoListModule(stores []buf.ModuleInputEntity) (*ProtoListModuleFormEntity, error) {
	form := &ProtoListModuleFormEntity{
		stores: stores,
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	return form, nil
}
