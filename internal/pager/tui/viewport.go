package tui

import "strings"

// VisibleLines returns the lines visible in the current viewport.
func VisibleLines(lines []string, top, height int) []string {
	if len(lines) == 0 || height <= 0 {
		return nil
	}

	top = clampTop(len(lines), top, height)
	end := top + height
	if end > len(lines) {
		end = len(lines)
	}

	return lines[top:end]
}

// RenderLines renders the visible slice of lines for the viewport.
func RenderLines(lines []string, top, height int) string {
	visible := VisibleLines(lines, top, height)
	if len(visible) == 0 {
		return ""
	}
	return strings.Join(visible, "\n")
}

func clampTop(lineCount, top, height int) int {
	if lineCount <= 0 {
		return 0
	}
	if height <= 0 {
		height = 1
	}
	maxTop := lineCount - height
	if maxTop < 0 {
		maxTop = 0
	}
	if top < 0 {
		return 0
	}
	if top > maxTop {
		return maxTop
	}
	return top
}
