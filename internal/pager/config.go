package pager

import (
	"os"
	"strings"
)

// GYATNoPager returns true if the GYAT_NO_PAGER environment variable is set
// to a truthy value. Recognized truthy values: 1, t, true, yes, on (case-insensitive).
func GYATNoPager() bool {
	v := os.Getenv("GYAT_NO_PAGER")
	if v == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "t", "true", "yes", "on":
		return true
	default:
		return false
	}
}
