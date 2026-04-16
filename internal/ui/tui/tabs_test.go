package tui

import (
	"testing"

	uiModel "github.com/refansa/gyat/v2/internal/ui/model"
)

func TestMoveActiveTabCyclesForwardAndBackward(t *testing.T) {
	t.Parallel()

	entry := uiModel.RepositoryEntry{
		StatusView: uiModel.RepositoryStatusView{
			ActiveTab: "overview",
			Tabs: []uiModel.StatusTab{
				{ID: "overview", Title: "Overview"},
				{ID: "staged", Title: "Staged"},
				{ID: "unstaged", Title: "Unstaged"},
			},
		},
	}

	moveActiveTab(&entry, 1)
	if entry.StatusView.ActiveTab != "staged" {
		t.Fatalf("active tab = %q, want staged", entry.StatusView.ActiveTab)
	}

	moveActiveTab(&entry, -1)
	if entry.StatusView.ActiveTab != "overview" {
		t.Fatalf("active tab = %q, want overview", entry.StatusView.ActiveTab)
	}

	moveActiveTab(&entry, -1)
	if entry.StatusView.ActiveTab != "unstaged" {
		t.Fatalf("active tab = %q, want unstaged", entry.StatusView.ActiveTab)
	}
}

func TestNormalizeEntryAddsDefaultOverviewTab(t *testing.T) {
	t.Parallel()

	entry := normalizeEntry(uiModel.RepositoryEntry{})
	if len(entry.StatusView.Tabs) != 1 {
		t.Fatalf("tabs len = %d, want 1", len(entry.StatusView.Tabs))
	}
	if entry.StatusView.ActiveTab != "overview" {
		t.Fatalf("active tab = %q, want overview", entry.StatusView.ActiveTab)
	}
}
