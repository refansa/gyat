package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelNavigationAndStatus(t *testing.T) {
	content := []byte("one\ntwo\nthree\nfour\nfive\n")
	m := newModel(content, 80, 4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(model)
	if m.top != 1 {
		t.Fatalf("top = %d, want 1", m.top)
	}

	view := m.View()
	if !strings.Contains(view, "2/5 40%") {
		t.Fatalf("expected status indicator in view, got:\n%s", view)
	}
}

func TestModelSearchAndHelp(t *testing.T) {
	content := []byte("alpha\nbeta\ngamma\nbeta again\n")
	m := newModel(content, 80, 6)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	if m.mode != searchMode {
		t.Fatalf("mode = %v, want searchMode", m.mode)
	}
	if cmd == nil {
		t.Fatal("expected focus command when entering search mode")
	}

	m.searchInput.SetValue("beta")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if len(m.matches) == 0 || m.matches[0] != 1 {
		t.Fatalf("unexpected matches after search: %v", m.matches)
	}
	if len(m.matches) != 2 {
		t.Fatalf("matches = %v, want 2 results", m.matches)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(model)
	if !m.helpVisible {
		t.Fatal("expected help to be visible")
	}
	if !strings.Contains(m.View(), "gyat viewer help") {
		t.Fatalf("expected help view, got:\n%s", m.View())
	}
}
