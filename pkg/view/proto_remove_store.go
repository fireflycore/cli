package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/buf"
	"strings"
)

type ProtoRemoveStoreFormEntity struct {
	problemIndex int

	Mode       string
	modeIndex  int
	Store      string
	storeIndex int

	module []buf.ModuleInputEntity
	local  []buf.LocalInputEntity
	modes  []string
}

func (model *ProtoRemoveStoreFormEntity) Init() tea.Cmd {
	return nil
}

func (model *ProtoRemoveStoreFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if model.problemIndex == 1 && model.storeIndex > 0 {
				model.storeIndex--
			}
		case "down":
			if model.problemIndex == 0 && (model.modeIndex < len(model.modes)-1) {
				model.modeIndex++
			}
			if model.problemIndex == 1 && model.Mode == "module" && (model.modeIndex < len(model.module)-1) {
				model.storeIndex++
			}
			if model.problemIndex == 1 && model.Mode == "local" && (model.modeIndex < len(model.local)-1) {
				model.storeIndex++
			}
		case "enter":
			switch model.problemIndex {
			case 0:
				model.Mode = strings.ToLower(model.modes[model.modeIndex])
				model.problemIndex++
			case 1:
				if model.Mode == "module" {
					if len(model.module) == 0 {
						return model, tea.Quit
					}
					model.Store = model.module[model.storeIndex].Module
				}
				if model.Mode == "local" {
					if len(model.local) == 0 {
						return model, tea.Quit
					}
					model.Store = model.local[model.storeIndex].Directory
				}
				model.problemIndex++
			}
			if model.problemIndex+1 > len(PROTO_REMOVE_STORE) {
				return model, tea.Quit
			}
		}
	}

	return model, cmd
}

func (model *ProtoRemoveStoreFormEntity) View() string {
	var str strings.Builder

	prefix := PrimaryColor.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_REMOVE_STORE[0])))

		for ii, item := range model.modes {
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
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_REMOVE_STORE[0])), PrimaryColor.Render(model.Mode)))
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(PROTO_REMOVE_STORE[1])))

		switch model.Mode {
		case "module":
			for ii, item := range model.module {
				selected := "  "
				if model.storeIndex == ii {
					selected = FocusColor.Render("->")
					item.Module = FocusColor.Render(item.Module)
				}
				str.WriteString(fmt.Sprintf("%s %s\n", selected, item.Module))
			}
		case "local":
			for ii, item := range model.local {
				selected := "  "
				if model.storeIndex == ii {
					selected = FocusColor.Render("->")
					item.Directory = FocusColor.Render(item.Directory)
				}
				str.WriteString(fmt.Sprintf("%s %s\n", selected, item.Directory))
			}
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(PROTO_REMOVE_STORE) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_REMOVE_STORE[0])), PrimaryColor.Render(model.Mode)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", PROTO_REMOVE_STORE[1])), PrimaryColor.Render(model.Store)))
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

func NewProtoRemoveStore(module []buf.ModuleInputEntity, local []buf.LocalInputEntity) (*ProtoRemoveStoreFormEntity, error) {
	form := &ProtoRemoveStoreFormEntity{
		module: module,
		local:  local,
		modes:  mode,
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
