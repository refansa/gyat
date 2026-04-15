package pager

import (
	"io"
	"os"
)

// IsTerminal reports whether the provided writer is connected to a terminal
// (a character device). It returns false for non-*os.File writers.
//
// This function intentionally mirrors the heuristic used elsewhere in the
// project: check for os.File and whether it's a character device. It is
// conservative and avoids platform-specific syscalls.
func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
