package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

type model struct {
	title       string
	entries     []uiModel.RepositoryEntry
	rows        []listRow
	selected    int
	top         int
	width       int
	height      int
	helpVisible bool
	quitting    bool
}

// NewModel builds the Bubble Tea model for the repository browser.
func NewModel(title string, entries []uiModel.RepositoryEntry, width, height int) tea.Model {
	return newModel(title, entries, width, height)
}

func newModel(title string, entries []uiModel.RepositoryEntry, width, height int) model {
	normalized := make([]uiModel.RepositoryEntry, len(entries))
	for index, entry := range entries {
		normalized[index] = normalizeEntry(entry)
	}

	m := model{
		title:    title,
		entries:  normalized,
		rows:     buildListRows(normalized),
		selected: firstSelectableEntry(normalized),
		width:    max(width, 40),
		height:   max(height, 8),
	}
	m.ensureVisible()
	return m
}

func firstSelectableEntry(entries []uiModel.RepositoryEntry) int {
	if len(entries) == 0 {
		return -1
	}
	return 0
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(msg.Width, 40)
		m.height = max(msg.Height, 8)
		m.ensureVisible()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", keyQuit, "esc":
			m.quitting = true
			return m, tea.Quit
		case keyHelp, keyHelpAlt:
			m.helpVisible = !m.helpVisible
			return m, nil
		case "j", keyDown:
			m.moveSelection(1)
		case "k", keyUp:
			m.moveSelection(-1)
		case " ", keyPageDown:
			m.pageSelection(1)
		case "b", keyPageUp:
			m.pageSelection(-1)
		case keyHome:
			m.selected = firstSelectableEntry(m.entries)
		case keyEnd:
			if len(m.entries) > 0 {
				m.selected = len(m.entries) - 1
			}
		case keyNextTab, keyRight:
			m.moveTab(1)
		case keyPrevTab, keyLeft:
			m.moveTab(-1)
		}
		m.ensureVisible()
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.helpVisible {
		return strings.Join([]string{
			renderHelpView(m.title),
			renderStatusLine(m.selected, len(m.entries), m.activeTabID()),
		}, "\n")
	}

	bodyHeight := m.viewportHeight()
	leftWidth := clamp(m.width/4, 22, 36)
	rightWidth := max(m.width-leftWidth-1, 24)
	topHeight := max((bodyHeight*2)/3, 8)
	bottomHeight := max(bodyHeight-topHeight-1, 4)
	leftTopHeight := max(bodyHeight-8, 6)
	leftBottomHeight := max(bodyHeight-leftTopHeight-1, 3)

	var selected *uiModel.RepositoryEntry
	if m.selected >= 0 && m.selected < len(m.entries) {
		selected = &m.entries[m.selected]
	}

	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		renderSidebar(m.entries, m.rows, m.selected, m.top, leftTopHeight, leftWidth),
		renderMiniPanel("other", []string{"contacts"}, leftWidth, leftBottomHeight),
	)

	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		renderDetail(selected, rightWidth, topHeight),
		renderInputPanel("Input", "Select a repository and switch tabs to inspect details.", rightWidth, bottomHeight),
	)

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		" ",
		rightColumn,
	)

	parts := []string{
		renderHeader(m.title, m.activeTabID()),
		body,
		renderStatusLine(m.selected, len(m.entries), m.activeTabID()),
	}
	return strings.Join(parts, "\n")
}

func (m *model) moveSelection(delta int) {
	if len(m.entries) == 0 {
		m.selected = -1
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.entries) {
		m.selected = len(m.entries) - 1
	}
}

func (m *model) pageSelection(direction int) {
	step := max(m.viewportHeight()/2, 1)
	m.moveSelection(direction * step)
}

func (m *model) moveTab(delta int) {
	if m.selected < 0 || m.selected >= len(m.entries) {
		return
	}
	moveActiveTab(&m.entries[m.selected], delta)
}

func (m *model) ensureVisible() {
	if len(m.rows) == 0 {
		m.top = 0
		return
	}

	selectedRow := selectedRowIndex(m.rows, m.selected)
	viewHeight := m.viewportHeight()
	if selectedRow < m.top {
		m.top = selectedRow
	}
	if selectedRow >= m.top+viewHeight {
		m.top = selectedRow - viewHeight + 1
	}
	maxTop := len(m.rows) - viewHeight
	if maxTop < 0 {
		maxTop = 0
	}
	if m.top > maxTop {
		m.top = maxTop
	}
	if m.top < 0 {
		m.top = 0
	}
}

func (m model) viewportHeight() int {
	visible := m.height - 2
	if visible < 1 {
		return 1
	}
	return visible
}

func (m model) activeTabID() string {
	if m.selected < 0 || m.selected >= len(m.entries) {
		return "overview"
	}
	return m.entries[m.selected].StatusView.Normalize().ActiveTab
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
