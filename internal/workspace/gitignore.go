package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/refansa/gyat/v2/internal/manifest"
)

const (
	gitIgnoreFileName     = ".gitignore"
	managedBlockStartLine = "# BEGIN gyat managed"
	managedBlockEndLine   = "# END gyat managed"
)

// SyncGitIgnore reconciles the gyat-managed block in the root .gitignore file.
// User-managed content outside the marked block is preserved.
func SyncGitIgnore(dir string, file manifest.File) (bool, error) {
	path := filepath.Join(dir, gitIgnoreFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("read %s: %w", gitIgnoreFileName, err)
		}
		if len(managedIgnoreEntries(file)) == 0 {
			return false, nil
		}
	}

	existing := string(data)
	next, err := syncGitIgnoreContent(existing, managedIgnoreEntries(file))
	if err != nil {
		return false, err
	}
	if next == existing {
		return false, nil
	}

	if next == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("remove %s: %w", gitIgnoreFileName, err)
		}
		return true, nil
	}

	if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", gitIgnoreFileName, err)
	}

	return true, nil
}

func managedIgnoreEntries(file manifest.File) []string {
	seen := map[string]struct{}{}
	entries := make([]string, 0, len(file.Ignore)+len(file.Repos))

	for _, pattern := range file.Ignore {
		pattern = normalizeIgnorePattern(pattern)
		if pattern == "" {
			continue
		}
		if _, exists := seen[pattern]; exists {
			continue
		}
		seen[pattern] = struct{}{}
		entries = append(entries, pattern)
	}

	for _, repo := range file.Repos {
		pattern := repoIgnorePattern(repo.Path)
		if pattern == "" {
			continue
		}
		if _, exists := seen[pattern]; exists {
			continue
		}
		seen[pattern] = struct{}{}
		entries = append(entries, pattern)
	}

	sort.Strings(entries)
	return entries
}

func normalizeIgnorePattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	pattern = strings.ReplaceAll(pattern, `\`, "/")
	return pattern
}

func repoIgnorePattern(path string) string {
	path = strings.TrimSpace(path)
	path = filepath.ToSlash(path)
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	return "/" + path + "/"
}

func syncGitIgnoreContent(existing string, managedEntries []string) (string, error) {
	newline := detectNewline(existing)
	lines := splitLines(existing)

	start := -1
	end := -1
	for index, line := range lines {
		switch line {
		case managedBlockStartLine:
			if start != -1 {
				return "", fmt.Errorf("malformed %s: multiple gyat managed blocks", gitIgnoreFileName)
			}
			start = index
		case managedBlockEndLine:
			if end != -1 {
				return "", fmt.Errorf("malformed %s: multiple gyat managed block terminators", gitIgnoreFileName)
			}
			end = index
		}
	}

	if (start == -1) != (end == -1) {
		return "", fmt.Errorf("malformed %s: incomplete gyat managed block", gitIgnoreFileName)
	}
	if start != -1 && end < start {
		return "", fmt.Errorf("malformed %s: gyat managed block terminator appears before block start", gitIgnoreFileName)
	}

	var prefix []string
	var suffix []string
	if start == -1 {
		prefix = trimTrailingBlankLines(lines)
	} else {
		prefix = trimTrailingBlankLines(lines[:start])
		suffix = trimLeadingBlankLines(lines[end+1:])
	}

	result := make([]string, 0, len(prefix)+len(suffix)+len(managedEntries)+4)
	result = append(result, prefix...)

	if len(managedEntries) > 0 {
		if len(result) > 0 {
			result = append(result, "")
		}
		result = append(result, managedBlockStartLine)
		result = append(result, managedEntries...)
		result = append(result, managedBlockEndLine)
	}

	if len(suffix) > 0 {
		if len(result) > 0 {
			result = append(result, "")
		}
		result = append(result, suffix...)
	}

	result = trimTrailingBlankLines(result)
	if len(result) == 0 {
		return "", nil
	}

	return strings.Join(result, newline) + newline, nil
}

func detectNewline(text string) string {
	if strings.Contains(text, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func trimLeadingBlankLines(lines []string) []string {
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	return lines
}

func trimTrailingBlankLines(lines []string) []string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
