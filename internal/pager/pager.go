package pager

import (
	"fmt"
	"io"
	"strings"
)

// Pager provides a simple interactive pager session. It supports basic
// navigation (page forward/back, line up/down) and search/match navigation.
// This is a lightweight implementation suitable for Phase 3 initial work.
type Pager struct {
	out         io.Writer
	lines       []string
	height      int // viewport height in lines
	cursorLine  int // 0-based index of first displayed line
	Matches     []int
	SearchQuery string
	closed      bool
}

// NewPager creates a new Pager that writes to the provided writer. The
// viewport height defaults to 24 lines; callers may adjust Pager.height
// directly for testing or specialized behavior.
func NewPager(w io.Writer) *Pager {
	return &Pager{out: w, height: 24}
}

// Render initializes the pager with content, writes the first page to the
// underlying writer and returns the number of bytes written. Subsequent
// navigation is handled via HandleKey and Seek.
func (p *Pager) Render(content []byte) (int, error) {
	if p == nil {
		return 0, fmt.Errorf("nil pager")
	}
	s := string(content)
	// Normalize line endings and split into lines
	s = strings.ReplaceAll(s, "\r\n", "\n")
	p.lines = strings.Split(s, "\n")
	if len(p.lines) > 0 && p.lines[len(p.lines)-1] == "" {
		// Trim trailing empty line introduced by a trailing newline
		p.lines = p.lines[:len(p.lines)-1]
	}
	p.cursorLine = 0
	p.closed = false
	return p.renderPage()
}

// renderPage writes the current page (from cursorLine) to the writer.
func (p *Pager) renderPage() (int, error) {
	if p == nil {
		return 0, fmt.Errorf("nil pager")
	}
	if p.out == nil {
		return 0, fmt.Errorf("nil writer")
	}
	if p.cursorLine < 0 {
		p.cursorLine = 0
	}
	if p.cursorLine >= len(p.lines) {
		// Nothing to write
		return 0, nil
	}
	end := p.cursorLine + p.height
	if end > len(p.lines) {
		end = len(p.lines)
	}
	total := 0
	for i := p.cursorLine; i < end; i++ {
		n, err := fmt.Fprintln(p.out, p.lines[i])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// HandleKey processes a single key input and updates the viewport accordingly.
// Supported keys: 'q' (quit), ' ' (page forward), 'b' (page back), 'j'/'k'
// (line down/up), 'n' (next match), 'N' (previous match).
func (p *Pager) HandleKey(r rune) (int, error) {
	if p == nil {
		return 0, fmt.Errorf("nil pager")
	}
	if p.closed {
		return 0, nil
	}
	switch r {
	case 'q':
		p.closed = true
		return 0, nil
	case ' ':
		p.cursorLine += p.height
		if p.cursorLine >= len(p.lines) {
			p.cursorLine = len(p.lines) - p.height
			if p.cursorLine < 0 {
				p.cursorLine = 0
			}
		}
		return p.renderPage()
	case 'b':
		p.cursorLine -= p.height
		if p.cursorLine < 0 {
			p.cursorLine = 0
		}
		return p.renderPage()
	case 'j':
		if p.cursorLine < len(p.lines)-1 {
			p.cursorLine++
		}
		return p.renderPage()
	case 'k':
		if p.cursorLine > 0 {
			p.cursorLine--
		}
		return p.renderPage()
	case 'n':
		return p.nextMatch()
	case 'N':
		return p.prevMatch()
	default:
		// Unsupported key — no-op
		return 0, nil
	}
}

// Search scans the content for occurrences of query and records match line
// indexes. It returns the number of matches found.
func (p *Pager) Search(query string) int {
	p.SearchQuery = query
	p.Matches = p.Matches[:0]
	if query == "" {
		return 0
	}
	for i, line := range p.lines {
		if strings.Contains(line, query) {
			p.Matches = append(p.Matches, i)
		}
	}
	return len(p.Matches)
}

// Seek moves the viewport to the matched line at matches[index] and renders the page.
func (p *Pager) Seek(index int) (int, error) {
	if index < 0 || index >= len(p.Matches) {
		return 0, fmt.Errorf("match index out of range")
	}
	p.cursorLine = p.Matches[index]
	return p.renderPage()
}

func (p *Pager) nextMatch() (int, error) {
	if len(p.Matches) == 0 {
		return 0, nil
	}
	// Find the first match index greater than or equal to cursorLine
	idx := 0
	for i, m := range p.Matches {
		if m > p.cursorLine {
			idx = i
			break
		}
		// if at end, wrap
		if i == len(p.Matches)-1 {
			idx = 0
		}
	}
	return p.Seek(idx)
}

func (p *Pager) prevMatch() (int, error) {
	if len(p.Matches) == 0 {
		return 0, nil
	}
	// Find the last match index less than cursorLine
	idx := len(p.Matches) - 1
	for i := len(p.Matches) - 1; i >= 0; i-- {
		if p.Matches[i] < p.cursorLine {
			idx = i
			break
		}
		if i == 0 {
			idx = len(p.Matches) - 1
		}
	}
	return p.Seek(idx)
}
