package pager

import (
	"bytes"
	"strings"
	"testing"
)

func BenchmarkPager_RenderSmall(b *testing.B) {
	content := strings.Repeat("line\n", 100)
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		buf.Reset()
		p := NewPager(&buf)
		p.height = 24
		_, _ = p.Render([]byte(content))
	}
}

func BenchmarkPager_RenderLarge(b *testing.B) {
	lines := make([]string, 0, 10000)
	for i := 0; i < 10000; i++ {
		lines = append(lines, "line")
	}
	content := strings.Join(lines, "\n") + "\n"
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		buf.Reset()
		p := NewPager(&buf)
		p.height = 40
		_, _ = p.Render([]byte(content))
	}
}

func BenchmarkDetectIsText_Text(b *testing.B) {
	s := []byte(strings.Repeat("hello world\n", 1000))
	for i := 0; i < b.N; i++ {
		_ = DetectIsText(s)
	}
}
