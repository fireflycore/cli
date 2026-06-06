package view

import "github.com/charmbracelet/lipgloss"

var (
	// PrimaryColor 用于主要信息和已确认内容。
	PrimaryColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	// WarningColor 用于底部操作提示。
	WarningColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	// DangerColor 用于错误或危险提示。
	DangerColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#e20000"))
	// InfoColor 用于普通问题文本。
	InfoColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	// FocusColor 用于当前焦点选项。
	FocusColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#009185"))
)
