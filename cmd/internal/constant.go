package internal

import "github.com/charmbracelet/lipgloss"

const CliName = "firefly"

var ASK = []string{
	"Please input your project name.",
	"Please choose your development language.",
	"Please select the database you want.",
}

var LANGUAGE = []string{
	"Go",
	"Rust",
	"NodeJS",
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
	Database["Go"] = []*DatabaseEntity{
		{Type: "Mysql", Name: "", Url: "", Select: false},
		{Type: "Mongo", Name: "", Url: "", Select: false},
		{Type: "Redis", Name: "", Url: "", Select: false},
	}
}
