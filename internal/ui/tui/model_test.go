package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

func TestModelNavigationAndVisibility(t *testing.T) {
	t.Parallel()

	entries := make([]uiModel.RepositoryEntry, 0, 6)
	for i := range 6 {
		entries = append(entries, uiModel.RepositoryEntry{
			ID:           string(rune('a' + i)),
			DisplayName:  "repo-" + string(rune('a'+i)),
			SummaryState: "clean",
			Metadata:     map[string]string{"group": "Group"},
			StatusView:   uiModel.RepositoryStatusView{Tabs: []uiModel.StatusTab{{ID: "overview", Title: "Overview", Content: "ok"}}},
		})
	}

	m := newModel("gyat list", entries, 80, 8)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(model)
	if m.selected != 1 {
		t.Fatalf("selected = %d, want 1", m.selected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(model)
	if m.selected <= 1 {
		t.Fatalf("expected page down to move selection, got %d", m.selected)
	}

	row := selectedRowIndex(m.rows, m.selected)
	if row < m.top || row >= m.top+m.viewportHeight() {
		t.Fatalf("selected row %d should remain visible within top=%d height=%d", row, m.top, m.viewportHeight())
	}
	if !strings.Contains(m.View(), "gyat list") {
		t.Fatalf("expected title in view, got:\n%s", m.View())
	}
}

func TestModelHelpToggle(t *testing.T) {
	t.Parallel()

	m := newModel("gyat list", nil, 80, 8)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(model)
	if !m.helpVisible {
		t.Fatal("expected help to be visible")
	}
	if !strings.Contains(m.View(), "gyat list help") {
		t.Fatalf("expected help view, got:\n%s", m.View())
	}
}

func TestModelViewShowsStructuredRegions(t *testing.T) {
	t.Parallel()

	entries := []uiModel.RepositoryEntry{{
		ID:           "repo-a",
		DisplayName:  "repo-a",
		SummaryState: "clean",
		Metadata:     map[string]string{"group": "Workspace"},
		StatusView: uiModel.RepositoryStatusView{Tabs: []uiModel.StatusTab{
			{ID: "overview", Title: "Overview", Content: []string{"hello"}},
			{ID: "staged", Title: "Staged", Content: []string{"none"}},
		}},
	}}

	m := newModel("gyat list", entries, 120, 30)
	view := m.View()
	for _, needle := range []string{"repo-a", "other", "contacts", "Input", "Overview"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected view to contain %q, got:\n%s", needle, view)
		}
	}
}
