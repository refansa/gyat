package pager

import (
	"bytes"
	"testing"
)

// Ensure DetectIsText treats binary data as non-text and that callers can
// safely write binary content directly when the helper signals non-text.
func TestDetectIsText_Binary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	binary := []byte{0x00, 0xff, 0x00, 0x10}
	if DetectIsText(binary) {
		t.Fatalf("expected binary to be detected as non-text")
	}

	// Simulate a caller writing binary directly when detection says non-text.
	if _, err := buf.Write(binary); err != nil {
		t.Fatalf("failed to write binary to buffer: %v", err)
	}
	if buf.Len() != len(binary) {
		t.Fatalf("unexpected buffer length: %d", buf.Len())
	}
}
