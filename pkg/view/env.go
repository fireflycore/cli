package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/store"
	"strings"
)

type EnvEchoFormEntity struct {
}

func (model *EnvEchoFormEntity) Init() tea.Cmd {
	return nil
}

func (model *EnvEchoFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return model, tea.Quit // 退出程序
		}
	}

	return model, cmd
}

func (model *EnvEchoFormEntity) View() string {
	var str strings.Builder

	str.WriteString(PrimaryColor.Render("-> Check the environment dependencies required by the current system.\n"))

	if len(store.Use.Buf.Version) != 0 {
		str.WriteString(FocusColor.Render(fmt.Sprintf("\n<- Buf cli version %s;\n", store.Use.Buf.Version)))
	} else {
		str.WriteString(WarningColor.Render("\n<- The current environment depends on buf cli v1.40.1 and above;\n"))
	}

	for index, item := range TIPS_TEXT {
		str.WriteString(fmt.Sprintf("\n%s %s", WarningColor.Render("->"), item))
		if index == len(TIPS_TEXT)-1 {
			str.WriteString("\n")
		}
	}

	return str.String()
}

func NewEnvEcho() (*EnvEchoFormEntity, error) {
	form := &EnvEchoFormEntity{}

	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	return form, nil
}
