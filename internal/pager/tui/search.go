package tui

import "strings"

// SearchLines returns the indexes of lines containing query.
func SearchLines(lines []string, query string) []int {
	if query == "" {
		return nil
	}

	matches := make([]int, 0)
	for index, line := range lines {
		if strings.Contains(line, query) {
			matches = append(matches, index)
		}
	}

	return matches
}

// NextMatch returns the next match line, wrapping at the end.
func NextMatch(matches []int, currentTop int) int {
	if len(matches) == 0 {
		return -1
	}
	for _, match := range matches {
		if match > currentTop {
			return match
		}
	}
	return matches[0]
}

// PrevMatch returns the previous match line, wrapping at the start.
func PrevMatch(matches []int, currentTop int) int {
	if len(matches) == 0 {
		return -1
	}
	for index := len(matches) - 1; index >= 0; index-- {
		if matches[index] < currentTop {
			return matches[index]
		}
	}
	return matches[len(matches)-1]
}

func matchOrdinal(matches []int, line int) int {
	for index, match := range matches {
		if match == line {
			return index
		}
	}
	return -1
}
