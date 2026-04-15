package pager

import (
	"bytes"
	"strings"
	"testing"
)

func TestPager_RenderWritesFirstPage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewPager(&buf)
	p.height = 3
	content := strings.Join([]string{"one", "two", "three", "four", "five"}, "\n") + "\n"
	n, err := p.Render([]byte(content))
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if n == 0 {
		t.Fatalf("expected non-zero bytes written")
	}
	got := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(got) != 3 || got[0] != "one" || got[2] != "three" {
		t.Fatalf("unexpected first page output: %v", got)
	}
}

func TestPager_NavigationSpaceAndBack(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewPager(&buf)
	p.height = 2
	content := strings.Join([]string{"a", "b", "c", "d", "e"}, "\n") + "\n"
	_, _ = p.Render([]byte(content))
	// page forward
	_, err := p.HandleKey(' ')
	if err != nil {
		t.Fatalf("page forward failed: %v", err)
	}
	// page back
	_, err = p.HandleKey('b')
	if err != nil {
		t.Fatalf("page back failed: %v", err)
	}
	// buffer should now contain three page renders (initial + forward + back)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) < 6 {
		t.Fatalf("unexpected output lines count: %d", len(lines))
	}
}

func TestPager_SearchAndSeek(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewPager(&buf)
	p.height = 2
	content := strings.Join([]string{"alpha", "beta", "gamma", "beta again", "delta"}, "\n") + "\n"
	_, _ = p.Render([]byte(content))
	matches := p.Search("beta")
	if matches != 2 {
		t.Fatalf("expected 2 matches for 'beta', got %d", matches)
	}
	// Seek to second match
	_, err := p.Seek(1)
	if err != nil {
		t.Fatalf("seek failed: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	// Seek rendered a page; ensure the last written page includes 'beta again'
	if !strings.Contains(lines[len(lines)-2], "beta again") && !strings.Contains(lines[len(lines)-1], "beta again") {
		t.Fatalf("expected 'beta again' to appear in last rendered page: %v", lines[len(lines)-3:])
	}
}
