package pager

import "io"

// OutputStream captures metadata about a command's output and provides
// helpers for detection (text vs binary, TTY, etc.).
type OutputStream struct {
	RawBytes []byte
	IsText   bool
	IsTTY    bool
	Writer   io.Writer
}

// NewOutputStream constructs an OutputStream given a writer and raw bytes.
// It detects whether the data appears textual and records the writer.
func NewOutputStream(w io.Writer, b []byte) *OutputStream {
	return &OutputStream{
		RawBytes: b,
		IsText:   DetectIsText(b),
		IsTTY:    IsTerminal(w),
		Writer:   w,
	}
}

// DetectIsText returns true if the byte slice appears to be text. The
// heuristic checks for NUL bytes and a high proportion of non-printable
// control characters to detect binary data. This is intentionally simple
// and conservative.
func DetectIsText(b []byte) bool {
	if len(b) == 0 {
		return true
	}

	// If there are NUL bytes, treat as binary
	for _, c := range b {
		if c == 0 {
			return false
		}
	}

	// Count control characters (except common whitespace: \n, \r, \t)
	var control int
	for _, c := range b {
		if c < 32 && c != 9 && c != 10 && c != 13 {
			control++
		}
	}

	// If more than 5% control characters, call it binary
	return float64(control)/float64(len(b)) < 0.05
}
