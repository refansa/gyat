---
description: "Review current gyat v2 test coverage, documentation, migration messaging, and release readiness"
agent: "plan"
---

# Review Gyat V2 Tests, Docs, and Rollout

Use this prompt to review what still needs to happen before gyat v2 can be
treated as stable. The manifest-based workspace rewrite already exists, so the
focus here is release readiness rather than greenfield planning.

## Current state

- Tests already use normal-repo fixtures in `t.TempDir()` rather than submodule fixtures.
- Command and integration coverage exists for the current workspace model.
- The remaining gaps are mostly around selector-aware behavior, migration edges, Windows validation, and stale docs.
- README and command help text are still catching up to the shipped v2 behavior.

## Relevant files

- [cmd/testhelper_test.go](../../cmd/testhelper_test.go)
- [cmd/integration_test.go](../../cmd/integration_test.go)
- [cmd/add_test.go](../../cmd/add_test.go)
- [cmd/status_test.go](../../cmd/status_test.go)
- [cmd/commit_test.go](../../cmd/commit_test.go)
- [cmd/rm_test.go](../../cmd/rm_test.go)
- [internal/git/git_test.go](../../internal/git/git_test.go)
- [README.md](../../README.md)
- [main.go](../../main.go)

## What to refine

1. The remaining coverage matrix for manifest creation, tracking, selection, status, commit, exec, sync, update, and migration.
2. Any fixture gaps that still make selector-aware or migration behavior hard to test.
3. Windows-specific validation gates.
4. README and CLI help sections that still describe pre-v2 or submodule-era behavior.
5. A release checklist for cutting the first stable v2 release.

## Deliverables

1. A test matrix for the remaining unit, integration, and migration coverage.
2. A short fixture review, including what is already good enough and what still needs work.
3. A docs rewrite checklist.
4. A rollout and migration communications plan.
5. A short list of residual risks that must be called out before release.

## Output format

### Test matrix
List the scenarios that must be covered before v2 is considered stable.

### Fixture review
Describe how test repositories are laid out today and what is still missing.

### Documentation checklist
List the docs and help text that need updates.

### Rollout plan
Describe how to stage, announce, and validate the v2 migration.

### Release blockers
List the issues that should block a v2 release if unresolved.