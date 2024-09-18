package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/buf"
	"strings"
)

type ProtoRemoveModuleFormEntity struct {
	problemIndex int

	Store       string
	storeIndex  int
	Module      string
	moduleIndex int

	stores []buf.ModuleInputEntity
	types  []string
}

func (model *ProtoRemoveModuleFormEntity) Init() tea.Cmd {
	return nil
}

func (model *ProtoRemoveModuleFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if model.problemIndex == 1 && model.moduleIndex > 0 {
				model.moduleIndex--
			}
		case "down":
			if model.problemIndex == 0 && (model.storeIndex < len(model.stores)-1) {
				model.storeIndex++
			}
			if model.problemIndex == 1 && (model.moduleIndex < len(model.types)-1) {
				model.moduleIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				store := model.stores[model.storeIndex]
				model.Store = store.Module
				model.types = store.Types
				model.problemIndex++
			case 1:
				model.Module = model.types[model.moduleIndex]
				model.problemIndex++
			}
			if model.problemIndex+1 > len(PROTO_REMOVE_MODULE) {
				return model, tea.Quit
			}
		}
	}

	return model, cmd
}

func (model *ProtoRemoveModuleFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_REMOVE_MODULE[0])))

		for ii, item := range model.stores {
			selected := "  "
			if model.storeIndex == ii {
				selected = FocusColor.Render("->")
				item.Module = FocusColor.Render(item.Module)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item.Module))
		}
	}

	// 处理第二个问题（如果尚未回答）
	if model.problemIndex == 1 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_REMOVE_MODULE[0])), PrimaryColor.Render(model.Store)))
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_REMOVE_MODULE[1])))
		for ii, item := range model.types {
			selected := "  "
			if model.moduleIndex == ii {
				selected = FocusColor.Render("->")
				item = FocusColor.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(PROTO_REMOVE_MODULE) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_REMOVE_MODULE[0])), PrimaryColor.Render(model.Store)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_REMOVE_MODULE[1])), PrimaryColor.Render(model.Module)))
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

func NewProtoRemoveModule(stores []buf.ModuleInputEntity) (*ProtoRemoveModuleFormEntity, error) {
	form := &ProtoRemoveModuleFormEntity{
		stores: stores,
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
