package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var CREATE_PROBLEM = []string{
	"Please input your project name.",
	"Please choose your development language.",
	"Please select the database you want.",
}

var Database = map[string][]*DatabaseEntity{}
var DatabaseList []*DatabaseEntity

var REGISTER = []string{
	"Etcd",
}

var (
	primary = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	warning = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	focus   = lipgloss.NewStyle().Foreground(lipgloss.Color("#009185"))
	info    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
)

func init() {
	Database["go"] = []*DatabaseEntity{
		{Type: "Mysql", DB: "", Url: "", Select: false},
		{Type: "Mongo", DB: "", Url: "", Select: false},
		{Type: "Redis", DB: "", Url: "", Select: false},
	}
}
