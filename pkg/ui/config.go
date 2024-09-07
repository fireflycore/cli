package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/config"
	"github.com/fireflycore/cli/pkg/store"
	"strings"
)

type ConfigFormEntity struct {
	problemIndex      int
	textLanguageIndex int
}

func (model ConfigFormEntity) Init() tea.Cmd {
	return nil
}

func (model ConfigFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				store.Use.Config.Global.TextLanguage = strings.ToLower(config.TEXT_LANGUAGE[model.textLanguageIndex])
				_ = store.Use.Config.UpdateGlobalConfig()
				model.problemIndex++
			}
			if model.problemIndex+1 > len(config.CONFIG_PROBLEM) {
				return model, tea.Quit
			}
		}
	}

	return model, cmd
}

func (model ConfigFormEntity) View() string {
	var str strings.Builder

	prefix := primary.Render("<-")
	// 处理第一个问题（如果尚未回答）
	if model.problemIndex == 0 {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, info.Render(fmt.Sprintf("%s -", config.CONFIG_PROBLEM[0])), primary.Render(store.Use.Config.Global.TextLanguage)))

		for ii, item := range config.TEXT_LANGUAGE {
			selected := "  "
			if model.textLanguageIndex == ii {
				selected = focus.Render("->")
				item = focus.Render(item)
			}
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 如果已经完成了所有问题
	if model.problemIndex == len(config.CONFIG_PROBLEM) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, info.Render(fmt.Sprintf("%s -", config.CONFIG_PROBLEM[0])), primary.Render(store.Use.Config.Global.TextLanguage)))
	} else {
		prefix = warning.Render("->")
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

func NewConfig() error {
	form := ConfigFormEntity{
		problemIndex:      0,
		textLanguageIndex: 0,
	}

	for index, item := range config.TEXT_LANGUAGE {
		if store.Use.Config.Global.TextLanguage == item {
			form.textLanguageIndex = index
			break
		}
	}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
