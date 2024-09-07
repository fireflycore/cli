package config

import "github.com/charmbracelet/lipgloss"

var Color = &ColorConfig{
	Primary: lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")),
	Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00")),
	Danger:  lipgloss.NewStyle().Foreground(lipgloss.Color("#e20000")),
	Info:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
	Focus:   lipgloss.NewStyle().Foreground(lipgloss.Color("#009185")),
}
