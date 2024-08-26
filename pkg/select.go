package pkg

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SelectModelEntity struct {
	ask   string
	items []string
	index int

	Value string
}

var (
	primary = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	focus   = lipgloss.NewStyle().Foreground(lipgloss.Color("#009185"))
	info    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
)

func (model SelectModelEntity) Init() tea.Cmd {
	return nil // 初始时不执行任何命令
}

func (model SelectModelEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		case "up":
			if model.index > 0 {
				model.index--
			}
		case "down":
			if model.index < len(model.items)-1 {
				model.index++
			}
		case "enter":
			model.Value = model.items[model.index]
			model.items = []string{}
			return model, tea.Quit
		}
	}
	return model, nil
}

func (model SelectModelEntity) View() string {
	s := primary.Render("<-")
	if len(model.Value) == 0 {
		s = fmt.Sprintf("%s %s", s, info.Render(model.ask))
	} else {
		s = fmt.Sprintf("%s %s %s\n", s, info.Render(model.ask), primary.Render(model.Value))
	}
	for i, item := range model.items {
		selected := "  "
		if model.index == i {
			selected = focus.Render("->")
			item = focus.Render(item)
		}
		s += fmt.Sprintf("\n%s %s", selected, item)
	}
	return s
}

func NewSelect(ask string, items []string) *tea.Program {
	repo := SelectModelEntity{
		ask:   ask,
		items: items,
	}

	return tea.NewProgram(repo)
}
