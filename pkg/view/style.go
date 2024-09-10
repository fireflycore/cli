package view

import "github.com/charmbracelet/lipgloss"

var (
	PrimaryColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	WarningColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	DangerColor  = lipgloss.NewStyle().Foreground(lipgloss.Color("#e20000"))
	InfoColor    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	FocusColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#009185"))
)
