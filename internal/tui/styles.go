package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("99")).
		Foreground(lipgloss.Color("231")).
		Padding(0, 1).
		MarginBottom(1)

	itemStyle = lipgloss.NewStyle().PaddingLeft(1)
	selectedItemStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("230")).
		PaddingLeft(1)

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)
)
