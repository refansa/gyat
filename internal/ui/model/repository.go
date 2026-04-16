package model

import "strings"

// RepositoryEntry is one repository shown in the interactive UI.
type RepositoryEntry struct {
	ID            string
	DisplayName   string
	Path          string
	CurrentBranch string
	SummaryState  string
	Metadata      map[string]string
	StatusView    RepositoryStatusView
}

// RepositoryStatusView describes the tabbed detail area for one repository.
type RepositoryStatusView struct {
	RepoID    string
	Tabs      []StatusTab
	ActiveTab string
}

// StatusTab is one selectable tab within the repository detail area.
type StatusTab struct {
	ID      string
	Title   string
	Content any
}

// Group returns the sidebar group label for the repository.
func (entry RepositoryEntry) Group() string {
	if entry.Metadata == nil {
		return "Other"
	}
	group := strings.TrimSpace(entry.Metadata["group"])
	if group == "" {
		return "Other"
	}
	return group
}

// Normalize ensures the status view always has a stable active tab.
func (view RepositoryStatusView) Normalize() RepositoryStatusView {
	if len(view.Tabs) == 0 {
		view.Tabs = []StatusTab{{
			ID:      "overview",
			Title:   "Overview",
			Content: "No details available.",
		}}
	}

	if strings.TrimSpace(view.ActiveTab) == "" {
		view.ActiveTab = view.Tabs[0].ID
	}

	for _, tab := range view.Tabs {
		if tab.ID == view.ActiveTab {
			return view
		}
	}

	view.ActiveTab = view.Tabs[0].ID
	return view
}
