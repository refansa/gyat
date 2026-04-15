package tui

import "testing"

func TestRenderLines(t *testing.T) {
	lines := []string{"one", "two", "three", "four"}
	if got := RenderLines(lines, 1, 2); got != "two\nthree" {
		t.Fatalf("RenderLines = %q", got)
	}
}

func TestVisibleLinesClampsBounds(t *testing.T) {
	lines := []string{"one", "two", "three"}
	visible := VisibleLines(lines, 99, 2)
	if len(visible) != 2 || visible[0] != "two" || visible[1] != "three" {
		t.Fatalf("unexpected visible lines: %v", visible)
	}
}
