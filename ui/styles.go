package ui

import "github.com/charmbracelet/lipgloss"

// inside here put the input and create the styleing with lipgloss

var (
	InputLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			Padding(1).
			MarginLeft(1)

	InputLabelMuteTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080")).
				MarginLeft(2).
				MarginRight(2).Padding(1)

	InputStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(1).
			Foreground(lipgloss.Color("#249D9F")).
			Bold(true)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f00"))

	DefaultMessage = lipgloss.NewStyle().Foreground(lipgloss.Color("#249D9F")).Padding(1)
	ErrorMessage   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f00"))
)
