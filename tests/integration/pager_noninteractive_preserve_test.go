package integration_test

import "testing"

func TestPagerNonInteractivePreservesOutputAndExitStatus(t *testing.T) {
	binary := buildGyatBinary(t)
	workspace := newIntegrationWorkspace(t)

	stdoutDefault, stderrDefault, exitDefault := runGyat(t, binary, workspace, nil, "status")
	stdoutEnv, stderrEnv, exitEnv := runGyat(t, binary, workspace, []string{"GYAT_NO_PAGER=1"}, "status")

	if exitDefault != 0 || exitEnv != 0 {
		t.Fatalf("expected zero exit codes, got default=%d env=%d", exitDefault, exitEnv)
	}
	if string(stdoutDefault) != string(stdoutEnv) {
		t.Fatalf("stdout differs when GYAT_NO_PAGER is set\ndefault:\n%s\nenv:\n%s", stdoutDefault, stdoutEnv)
	}
	if string(stderrDefault) != string(stderrEnv) {
		t.Fatalf("stderr differs when GYAT_NO_PAGER is set\ndefault:\n%s\nenv:\n%s", stderrDefault, stderrEnv)
	}
}
