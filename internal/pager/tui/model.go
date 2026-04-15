package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputMode int

const (
	normalMode inputMode = iota
	searchMode
)

type model struct {
	lines       []string
	top         int
	width       int
	height      int
	query       string
	matches     []int
	matchIndex  int
	mode        inputMode
	helpVisible bool
	quitting    bool
	searchInput textinput.Model
}

// NewModel builds the Bubble Tea model for the pager UI.
func NewModel(content []byte, width, height int) tea.Model {
	return newModel(content, width, height)
}

func newModel(content []byte, width, height int) model {
	input := textinput.New()
	input.Prompt = "/"
	input.Placeholder = "search"
	input.CharLimit = 256
	input.Width = max(width-2, 10)
	input.Blur()

	return model{
		lines:       splitLines(content),
		width:       width,
		height:      max(height, 3),
		matchIndex:  -1,
		searchInput: input,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height, 3)
		m.searchInput.Width = max(msg.Width-2, 10)
		m.normalizeTop()
		return m, nil

	case tea.KeyMsg:
		if m.mode == searchMode {
			return m.updateSearch(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case " ", "space":
			m.top += m.viewportHeight()
			m.normalizeTop()
		case "b":
			m.top -= m.viewportHeight()
			m.normalizeTop()
		case "j", "down":
			m.top++
			m.normalizeTop()
		case "k", "up":
			m.top--
			m.normalizeTop()
		case "n":
			m.jumpToMatch(NextMatch(m.matches, m.top))
		case "N":
			m.jumpToMatch(PrevMatch(m.matches, m.top))
		case "/":
			m.helpVisible = false
			m.mode = searchMode
			m.searchInput.SetValue(m.query)
			m.searchInput.CursorEnd()
			return m, m.searchInput.Focus()
		case "?", "h":
			m.helpVisible = !m.helpVisible
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.helpVisible {
		return strings.Join([]string{
			renderHelpView(),
			renderStatusLine(m.top, len(m.lines), m.query, m.matches, m.matchIndex),
		}, "\n")
	}

	body := RenderLines(m.lines, m.top, m.viewportHeight())
	if body == "" {
		body = renderEmptyBody()
	}

	parts := []string{body}
	if m.mode == searchMode {
		parts = append(parts, renderSearchPrompt(m.searchInput))
	}
	parts = append(parts, renderStatusLine(m.top, len(m.lines), m.query, m.matches, m.matchIndex))
	return strings.Join(parts, "\n")
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = normalMode
		m.searchInput.Blur()
		m.normalizeTop()
		return m, nil
	case "enter":
		m.mode = normalMode
		m.searchInput.Blur()
		m.query = m.searchInput.Value()
		m.matches = SearchLines(m.lines, m.query)
		m.matchIndex = -1
		if len(m.matches) > 0 {
			m.jumpToMatch(m.matches[0])
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m *model) jumpToMatch(line int) {
	if line < 0 {
		m.matchIndex = -1
		return
	}
	m.top = line
	m.normalizeTop()
	m.matchIndex = matchOrdinal(m.matches, line)
}

func (m *model) normalizeTop() {
	m.top = clampTop(len(m.lines), m.top, m.viewportHeight())
	if m.matchIndex >= 0 && m.matchIndex < len(m.matches) && m.matches[m.matchIndex] != m.top {
		m.matchIndex = matchOrdinal(m.matches, m.top)
	}
}

func (m model) viewportHeight() int {
	reserved := 1
	if m.mode == searchMode {
		reserved++
	}
	visible := m.height - reserved
	if visible < 1 {
		return 1
	}
	return visible
}

func splitLines(content []byte) []string {
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
