package integration_test

import "testing"

// This integration test validates piping behavior. It is intended to be run
// on a Windows runner with a real TTY/pipe environment. For automated runs
// the project already includes unit tests covering piping semantics; this
// placeholder documents the manual integration requirement.
func TestPagerPipeIntegration(t *testing.T) {
	t.Skip("manual integration test: run on a Windows runner with a TTY/pipes to validate piping behavior; see specs/001-windows-pager/quickstart.md")
}
