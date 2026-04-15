package pager

// Session represents the interactive state of a pager invocation.
// Fields will be expanded in Phase 3 when implementing navigation and
// search features.
type Session struct {
	ID             string
	CursorPosition int
	ViewportHeight int
	ViewportWidth  int
	SearchQuery    string
	Matches        []int
	// Mode indicates the current pager mode (e.g., "normal", "search").
	Mode string
}
