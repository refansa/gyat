package tui

import (
	"strconv"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

func BenchmarkNavigation200Repos(b *testing.B) {
	entries := make([]uiModel.RepositoryEntry, 0, 200)
	for i := range 200 {
		entries = append(entries, uiModel.RepositoryEntry{
			ID:           strconv.Itoa(i),
			DisplayName:  "repo-" + strconv.Itoa(i),
			SummaryState: "clean",
			Metadata:     map[string]string{"group": "Workspace"},
			StatusView:   uiModel.RepositoryStatusView{Tabs: []uiModel.StatusTab{{ID: "overview", Title: "Overview", Content: "ok"}}},
		})
	}

	m := newModel("gyat list", entries, 120, 30)
	b.ResetTimer()
	for range b.N {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(model)
	}
}
