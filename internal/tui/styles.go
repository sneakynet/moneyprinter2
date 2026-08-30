package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Padding(0, 1)

	greetingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Align(lipgloss.Center)

	itemStyle = lipgloss.NewStyle().PaddingLeft(1)
	selectedItemStyle = lipgloss.NewStyle().Reverse(true).PaddingLeft(1)

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	dialogStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	buttonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Padding(0, 2)

	buttonFocusedStyle = lipgloss.NewStyle().Reverse(true).Padding(0, 2)

	overlayStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))
)
