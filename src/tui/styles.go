package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#A78BFA", Dark: "#A78BFA"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	fadedWhite = lipgloss.AdaptiveColor{Light: "#B0B0B0", Dark: "#6B6B6B"}

	criticalColor = lipgloss.Color("#FF0055")
	highColor     = lipgloss.Color("#FF4400")
	mediumColor   = lipgloss.Color("#FFAA00")
	lowColor      = lipgloss.Color("#00AAFF")
	infoColor     = lipgloss.Color("#888888")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlight).
			Padding(0, 1)

	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	menuSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(lipgloss.Color("#000000")).
				Background(highlight)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle)

	helpStyle = lipgloss.NewStyle().
			Foreground(fadedWhite).
			Padding(0, 1)

	findingPathStyle = lipgloss.NewStyle().
				Foreground(infoColor)

	findingLineStyle = lipgloss.NewStyle().
				Foreground(subtle)

	statusBarStyle = lipgloss.NewStyle().
			Background(subtle).
			Foreground(lipgloss.Color("#000000")).
			Padding(1, 2)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(highlight)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0055"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(special).
			Padding(0, 1)
)

func severityStyle(severity string) lipgloss.Style {
	s := lipgloss.NewStyle().Bold(true)
	if severity == "critical" {
		return s.Foreground(criticalColor)
	}
	if severity == "high" {
		return s.Foreground(highColor)
	}
	if severity == "medium" {
		return s.Foreground(mediumColor)
	}
	if severity == "low" {
		return s.Foreground(lowColor)
	}
	return s
}

func severityBadge(severity string) string {
	return severityStyle(severity).Render(severity)
}
