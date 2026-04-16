package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

var (
	tabsActiveTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Padding(0, 1)
	tabsInactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	tabsDetailTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	tabsMetaStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	tabsPanelStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	tabsDimStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func NormalizeEntry(entry uiModel.RepositoryEntry) uiModel.RepositoryEntry {
	entry.StatusView = entry.StatusView.Normalize()
	return entry
}

func MoveActiveTab(entry *uiModel.RepositoryEntry, delta int) {
	if entry == nil {
		return
	}
	entry.StatusView = entry.StatusView.Normalize()
	if len(entry.StatusView.Tabs) == 0 {
		return
	}

	active := 0
	for index, tab := range entry.StatusView.Tabs {
		if tab.ID == entry.StatusView.ActiveTab {
			active = index
			break
		}
	}

	active = (active + delta) % len(entry.StatusView.Tabs)
	if active < 0 {
		active += len(entry.StatusView.Tabs)
	}
	entry.StatusView.ActiveTab = entry.StatusView.Tabs[active].ID
}

func RenderDetail(entry *uiModel.RepositoryEntry, width, height int) string {
	if width < 20 {
		width = 20
	}
	if height < 3 {
		height = 3
	}
	if entry == nil {
		return tabsPanelStyle.Width(width).Height(height + 1).Render(tabsDimStyle.Render("No repository selected."))
	}

	current := NormalizeEntry(*entry)
	tab := activeTab(current)

	tabLabels := make([]string, 0, len(current.StatusView.Tabs))
	for _, candidate := range current.StatusView.Tabs {
		label := tabsInactiveTabStyle.Render(candidate.Title)
		if candidate.ID == current.StatusView.ActiveTab {
			label = tabsActiveTabStyle.Render(candidate.Title)
		}
		tabLabels = append(tabLabels, label)
	}

	headline := tabsDetailTitleStyle.Render(TrimToWidth(current.DisplayName, width-2))
	subtitle := tabsMetaStyle.Render(TrimToWidth(fmt.Sprintf("%s | %s", current.Path, current.SummaryState), width-2))
	content := trimContent(tab.Content, width-2)
	body := strings.Join([]string{
		headline,
		subtitle,
		strings.Join(tabLabels, " "),
		"",
		content,
	}, "\n")

	return tabsPanelStyle.Width(width).Height(height + 1).Render(body)
}

func RenderInputPanel(title string, placeholder string, width, height int) string {
	if width < 20 {
		width = 20
	}
	if height < 3 {
		height = 3
	}

	body := strings.Join([]string{
		tabsDetailTitleStyle.Render(TrimToWidth(title, width-2)),
		"",
		tabsDimStyle.Render(TrimToWidth(placeholder, width-2)),
	}, "\n")

	return tabsPanelStyle.Width(width).Height(height + 1).Render(body)
}

func activeTab(entry uiModel.RepositoryEntry) uiModel.StatusTab {
	entry = NormalizeEntry(entry)
	for _, tab := range entry.StatusView.Tabs {
		if tab.ID == entry.StatusView.ActiveTab {
			return tab
		}
	}
	return entry.StatusView.Tabs[0]
}

func trimContent(content any, width int) string {
	text := stringifyContent(content)
	if text == "" {
		return tabsDimStyle.Render("No details available.")
	}
	lines := strings.Split(text, "\n")
	for index, line := range lines {
		lines[index] = TrimToWidth(line, width)
	}
	return strings.Join(lines, "\n")
}

func stringifyContent(content any) string {
	switch value := content.(type) {
	case nil:
		return ""
	case string:
		return value
	case []string:
		return strings.Join(value, "\n")
	default:
		return fmt.Sprint(value)
	}
}
