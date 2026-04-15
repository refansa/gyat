package integration_test

import "testing"

// This test documents that --no-pager and GYAT_NO_PAGER disable interactive
// paging. The behaviour is covered by unit tests in cmd/pager_test.go; this
// placeholder is kept for an integration harness if a Windows runner is used.
func TestNoPagerIntegration(t *testing.T) {
	t.Skip("manual integration test: verify --no-pager and GYAT_NO_PAGER behavior on Windows runner")
}
