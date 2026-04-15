package pager

import (
	"bytes"
	"testing"
)

func TestDetectIsText(t *testing.T) {
	t.Parallel()

	text := []byte("hello world\nthis is text\n")
	if !DetectIsText(text) {
		t.Fatalf("expected text to be detected as text")
	}

	binary := []byte{0x00, 0xff, 0x00, 0x10}
	if DetectIsText(binary) {
		t.Fatalf("expected binary to be detected as non-text")
	}
}

func TestNewOutputStream(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	b := []byte("sample")
	os := NewOutputStream(&buf, b)
	if !os.IsText {
		t.Fatalf("expected output stream to mark sample as text")
	}
	if os.IsTTY {
		t.Fatalf("expected non-file writer to not be a TTY")
	}
}
