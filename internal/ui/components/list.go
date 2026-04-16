package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

type ListRow struct {
	Header     string
	EntryIndex int
	Selectable bool
}

var (
	listGroupHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	listSelectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	listRowStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	listMetaStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	listPanelStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	listDimStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func BuildListRows(entries []uiModel.RepositoryEntry) []ListRow {
	rows := make([]ListRow, 0, len(entries)*2)
	seenGroups := map[string]bool{}
	for index, entry := range entries {
		group := entry.Group()
		if !seenGroups[group] {
			rows = append(rows, ListRow{Header: group})
			seenGroups[group] = true
		}
		rows = append(rows, ListRow{EntryIndex: index, Selectable: true})
	}
	return rows
}

func SelectedRowIndex(rows []ListRow, selectedEntry int) int {
	if selectedEntry < 0 {
		return 0
	}
	for index, row := range rows {
		if row.Selectable && row.EntryIndex == selectedEntry {
			return index
		}
	}
	return 0
}

func RenderSidebar(entries []uiModel.RepositoryEntry, rows []ListRow, selectedEntry, top, height, width int) string {
	if width < 12 {
		width = 12
	}
	if height < 1 {
		height = 1
	}

	visible := rows
	if top < len(rows) {
		visible = rows[top:]
	}
	if len(visible) > height {
		visible = visible[:height]
	}

	lines := make([]string, 0, len(visible))
	for _, row := range visible {
		if !row.Selectable {
			lines = append(lines, listGroupHeaderStyle.Render(TrimToWidth(row.Header, width-2)))
			continue
		}

		entry := entries[row.EntryIndex]
		line := TrimToWidth(entry.DisplayName, width-2)
		meta := TrimToWidth(entry.SummaryState, width-2)
		cell := listRowStyle.Render(line) + "\n" + listMetaStyle.Render(meta)
		if row.EntryIndex == selectedEntry {
			cell = listSelectedRowStyle.Width(width - 2).Render(line + "\n" + meta)
		}
		lines = append(lines, cell)
	}

	if len(lines) == 0 {
		lines = append(lines, listDimStyle.Render("(no repositories)"))
	}

	body := strings.Join(lines, "\n")
	return listPanelStyle.Width(width).Height(height + 1).Render(body)
}

func RenderMiniPanel(title string, lines []string, width, height int) string {
	if width < 12 {
		width = 12
	}
	if height < 1 {
		height = 1
	}
	if len(lines) == 0 {
		lines = []string{"(empty)"}
	}

	rendered := make([]string, 0, len(lines)+1)
	rendered = append(rendered, listGroupHeaderStyle.Render(TrimToWidth(title, width-2)))
	for _, line := range lines {
		rendered = append(rendered, listMetaStyle.Render(TrimToWidth(line, width-2)))
	}

	body := strings.Join(rendered, "\n")
	return listPanelStyle.Width(width).Height(height + 1).Render(body)
}

func TrimToWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= width {
		return string(runes)
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "…"
}
