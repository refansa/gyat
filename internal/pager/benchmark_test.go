package pager

import (
	"bytes"
	"strings"
	"testing"

	pagertui "github.com/refansa/gyat/v2/internal/pager/tui"
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
	lines := make([]string, 0, 50000)
	for i := 0; i < 50000; i++ {
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

func BenchmarkTUIModel_RenderLarge(b *testing.B) {
	lines := make([]string, 0, 50000)
	for i := 0; i < 50000; i++ {
		lines = append(lines, "line")
	}
	content := []byte(strings.Join(lines, "\n") + "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model := pagertui.NewModel(content, 120, 40)
		_ = model.View()
	}
}

func BenchmarkDetectIsText_Text(b *testing.B) {
	s := []byte(strings.Repeat("hello world\n", 1000))
	for i := 0; i < b.N; i++ {
		_ = DetectIsText(s)
	}
}
