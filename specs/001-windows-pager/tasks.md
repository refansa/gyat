# Tasks: windows-pager

**Input**: plan.md, spec.md, research.md

## Phase 1: Setup (Shared Infrastructure)

- [ ] T001 Create cross-platform pager module under `internal/pager` or `src/cli/pager`
- [ ] T002 Add entry points and CLI flag parsing for `--no-pager` and honor `GYAT_NO_PAGER`
- [ ] T003 Add configuration/feature-flag scaffolding if needed

## Phase 2: Implementation

- [ ] T010 Extract or reference existing Unix pager code as the canonical behavior
- [ ] T011 Implement Windows adaptation layer:
  - Key input capture (space, b, j/k, arrows, /, n, N, q)
  - Resize handling where supported
  - Search mode and highlight navigation
- [ ] T012 Ensure pager only activates when stdout is a TTY
- [ ] T013 Ensure binary output bypasses pager
- [ ] T014 Implement byte-for-byte preservation when pager bypassed

## Phase 3: Tests

- [ ] T020 Unit test: TTY detection logic (simulate tty vs non-tty)
- [ ] T021 Unit test: Binary data bypass
- [ ] T022 Integration test: Pipe output and assert receiving process gets unmodified data
- [ ] T023 Integration/manual: Interactive test in Windows Terminal verifying navigation and search

## Phase 4: Docs & Quickstart

- [ ] T030 Update CLI help text to document pager behavior and `--no-pager`
- [ ] T031 Add quickstart.md steps to contributor docs and release notes

## Phase 5: Release & Rollout

- [ ] T040 Add migration note to changelog if behavior differs from previous Windows implementation
- [ ] T041 Monitor for user reports and regressions post-release
