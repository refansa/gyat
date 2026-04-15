package tui

import "testing"

func TestSearchLines(t *testing.T) {
	lines := []string{"alpha", "beta", "beta again", "gamma"}
	matches := SearchLines(lines, "beta")
	if len(matches) != 2 || matches[0] != 1 || matches[1] != 2 {
		t.Fatalf("unexpected matches: %v", matches)
	}
}

func TestNextPrevMatchWrap(t *testing.T) {
	matches := []int{2, 5, 9}
	if got := NextMatch(matches, 5); got != 9 {
		t.Fatalf("NextMatch = %d, want 9", got)
	}
	if got := NextMatch(matches, 9); got != 2 {
		t.Fatalf("NextMatch wrap = %d, want 2", got)
	}
	if got := PrevMatch(matches, 5); got != 2 {
		t.Fatalf("PrevMatch = %d, want 2", got)
	}
	if got := PrevMatch(matches, 2); got != 9 {
		t.Fatalf("PrevMatch wrap = %d, want 9", got)
	}
}
