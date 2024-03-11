package ui

import "github.com/charmbracelet/lipgloss"

// inside here put the input and create the styleing with lipgloss

var (
	primaryColor   = lipgloss.Color("#249D9F")
	secondaryColor = lipgloss.Color("#F28030")
	muteColor      = lipgloss.Color("#808080")
	errorColor     = lipgloss.Color("#f00")
	successColor   = lipgloss.Color("#249D9F")

	InputLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			Padding(1).
			MarginLeft(1)

	InputLabelMuteTextStyle = lipgloss.NewStyle().
				Foreground(muteColor).
				MarginLeft(2).
				MarginRight(2).Padding(1)

	InputStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(1).
			Foreground(primaryColor).
			Bold(true)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor)

	DefaultMessage = lipgloss.NewStyle().Foreground(secondaryColor).MarginLeft(5).MarginTop(1)
	SuccessMessage = lipgloss.NewStyle().Foreground(successColor).MarginLeft(5).MarginTop(1)
	ErrorMessage   = lipgloss.NewStyle().Foreground(errorColor)
)
