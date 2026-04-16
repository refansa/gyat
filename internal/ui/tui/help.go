package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)
	helpTitleStyle = lipgloss.NewStyle().Bold(true)
	statusStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
)

func renderHelpView(title string) string {
	lines := []string{
		helpTitleStyle.Render(title + " help"),
		"up/down, j/k: move repository selection",
		"page up/down, home/end: jump through the list",
		"tab/shift-tab or left/right: switch tabs",
		"?/h: toggle this help",
		"q or esc: quit back to the shell",
	}
	return helpBoxStyle.Render(strings.Join(lines, "\n"))
}

func renderStatusLine(selected, total int, activeTab string) string {
	position := "0/0"
	if total > 0 {
		position = fmt.Sprintf("%d/%d", selected+1, total)
	}
	if activeTab == "" {
		activeTab = "overview"
	}
	return statusStyle.Render(position + " | tab: " + activeTab + " | q quit ? help")
}

func renderHeader(title string, activeTab string) string {
	if activeTab == "" {
		return renderTitle(title)
	}
	return renderTitle(title + " [" + activeTab + "]")
}

func renderTitle(title string) string {
	return titleStyle.Render(title)
}
