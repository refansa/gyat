package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	searchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
)

func renderHelpView() string {
	lines := []string{
		helpTitleStyle.Render("gyat viewer help"),
		"j/k or up/down: move one line",
		"space/b: page forward/back",
		"/: search for a substring",
		"n/N: jump to next/previous match",
		"?/h: toggle this help",
		"q: quit back to the shell",
	}
	return helpBoxStyle.Render(strings.Join(lines, "\n"))
}

func renderStatusLine(top, total int, query string, matches []int, matchIndex int) string {
	position := "0/0 0%"
	if total > 0 {
		line := top + 1
		if line > total {
			line = total
		}
		position = fmt.Sprintf("%d/%d %d%%", line, total, percentage(line, total))
	}

	search := "search: off"
	if query != "" {
		search = fmt.Sprintf("search: /%s (%d matches)", query, len(matches))
		if matchIndex >= 0 {
			search = fmt.Sprintf("search: /%s (%d/%d)", query, matchIndex+1, len(matches))
		}
	}

	return statusStyle.Render(position + " | " + search + " | q quit ? help")
}

func renderSearchPrompt(input textinput.Model) string {
	return searchStyle.Render(input.View())
}

func renderEmptyBody() string {
	return dimStyle.Render("(no output)")
}

func percentage(current, total int) int {
	if total <= 0 {
		return 0
	}
	return (current * 100) / total
}
