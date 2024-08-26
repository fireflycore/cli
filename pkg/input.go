package pkg

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputModelEntity struct {
	ask   string
	input textinput.Model

	Value string
}

func (model InputModelEntity) Init() tea.Cmd {
	return textinput.Blink
}

func (model InputModelEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return model, tea.Quit // 退出程序
		case "enter":
			model.Value = model.input.View()
			return model, tea.Quit
		}
	}

	model.input, cmd = model.input.Update(msg)

	return model, cmd
}

func (model InputModelEntity) View() string {
	s := primary.Render("<-")
	if len(model.Value) == 0 {
		s = fmt.Sprintf("%s %s", s, info.Render(model.ask))
	} else {
		s = fmt.Sprintf("%s %s %s", s, info.Render(fmt.Sprintf("%s -", model.ask)), primary.Render(model.Value))
	}
	s += fmt.Sprintf("\n-> %s", model.input.View())
	return s
}

func NewInput(ask string) *tea.Program {
	input := textinput.New()
	input.Prompt = ""
	input.Focus()

	repo := InputModelEntity{
		input: input,
		ask:   ask,
	}

	return tea.NewProgram(repo)
}
