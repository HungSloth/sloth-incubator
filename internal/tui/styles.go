package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	accentColor   = lipgloss.Color("#7C3AED") // purple
	successColor  = lipgloss.Color("#10B981") // green
	errorColor    = lipgloss.Color("#EF4444") // red
	mutedColor    = lipgloss.Color("#6B7280") // gray
	selectedColor = lipgloss.Color("#7C3AED") // purple

	// Header style
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(accentColor).
			Padding(0, 2).
			MarginBottom(1)

	// Title for header box
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	// Active list item
	activeItemStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Inactive list item
	inactiveItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D1D5DB"))

	// Muted text
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Success text
	successStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// Error text
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	// Selected/focused input
	focusedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Label style
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Width(20)

	// Value style for confirmation screen
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	// Help text at the bottom
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Border box
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(0, 1)

	// App wrapper
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)
)
