---

description: "Task list for feature implementation: windows-pager"
---

# Tasks: windows-pager

**Input**: plan.md, spec.md, research.md, data-model.md

## Phase 1: Setup (Shared Infrastructure)

Purpose: Create the shared pager package and small plumbing that other tasks depend on.

- [ ] T001 [P] Create internal/pager package skeleton with public API placeholders in internal/pager/pager.go, internal/pager/session.go, internal/pager/output.go, internal/pager/doc.go
 - [X] T001 [P] Create internal/pager package skeleton with public API placeholders in internal/pager/pager.go, internal/pager/session.go, internal/pager/output.go, internal/pager/doc.go

  Note: Skeleton files created (internal/pager/*.go). Interactive rendering and helpers to be implemented in subsequent tasks.

---

## Phase 2: Foundational (Blocking Prerequisites)

Purpose: Implement core types, environment/config handling and helpers that ALL user stories will use. These must be completed before story implementation.

- [X] T002 [P] Implement OutputStream model and helpers in internal/pager/output.go (fields: RawBytes, IsText, IsTTY) and add IsText detection helper in internal/pager/output.go
- [X] T003 [P] Add terminal detection helper IsTerminal(writer io.Writer) in internal/pager/tty.go (exported) and document behavior in internal/pager/doc.go
- [X] T004 [P] Add pager configuration helper GYATNoPager() in internal/pager/config.go that reads and normalizes GYAT_NO_PAGER env var
- [X] T005 [P] Update cmd/pager.go to call internal/pager.GYATNoPager() and ensure --no-pager flag and the env override both disable paging (file: cmd/pager.go)
- [X] T006 [P] Add unit tests for OutputStream, IsText detection, IsTerminal and GYATNoPager in internal/pager/*_test.go and augment cmd/pager_test.go (files: internal/pager/output.go, internal/pager/tty.go, internal/pager/config.go, cmd/pager_test.go)
- [ ] T007 [P] Update docs to record the new env var and flag behavior in specs/001-windows-pager/quickstart.md and add an entry to CHANGELOG.md (files: specs/001-windows-pager/quickstart.md, CHANGELOG.md)
 - [X] T007 [P] Update docs to record the new env var and flag behavior in specs/001-windows-pager/quickstart.md and add an entry to CHANGELOG.md (files: specs/001-windows-pager/quickstart.md, CHANGELOG.md)

---

## Phase 3: User Story 1 - Improved Windows paging (Priority: P1) 🎯 MVP

Goal: Provide an interactive pager on Windows that matches Unix-like pager behavior (navigation keys, search, resize where supported) while preserving stdout semantics when bypassed.

Independent Test: Launch a long-output command (e.g., `gyat list` against a large workspace) on Windows and verify interactive navigation (space, b, j/k, arrows), search with `/`, jump with `n`/`N`, and `q` to quit. When `--no-pager` or GYAT_NO_PAGER is set, output must stream directly to stdout.

### Tests for User Story 1

> OPTIONAL: Add automated tests where feasible; include a manual integration test plan for interactive verification.

- [ ] T010 [US1] Define PagerSession struct and in-memory session state in internal/pager/session.go (fields: id, CursorPosition, ViewportHeight, ViewportWidth, SearchQuery, Matches, Mode)
 - [X] T010 [US1] Define PagerSession struct and in-memory session state in internal/pager/session.go (fields: id, CursorPosition, ViewportHeight, ViewportWidth, SearchQuery, Matches, Mode)
 - [X] T011 [US1] Implement pager rendering and navigation engine in internal/pager/render.go and internal/pager/pager.go (methods: Render(content []byte), HandleKey(input rune), Seek(matchIndex int))
 - [X] T012 [US1] Implement Windows-specific input/resize adaptation in internal/pager/windows.go using build tags (//go:build windows) to capture key presses and resize events where available
 - [X] T013 [US1] Implement search mode and match navigation (commands: '/', 'n', 'N') in internal/pager/search.go and integrate with session state
 - [X] T014 [US1] Wire the internal/pager API into the CLI: replace direct pager exec paths in cmd/pager.go and cmd/status.go with calls to internal/pager.NewPagerSession / internal/pager.Render (files: cmd/pager.go, cmd/status.go)
- [ ] T015 [US1] Add an integration test or harness (can be manual-first) at tests/integration/pager_navigation_test.go that validates navigation and search behavior on a Windows runner or simulated TTY
- [ ] T016 [US1] Update specs/001-windows-pager/quickstart.md with explicit interactive validation steps and examples to reproduce acceptance scenarios (file: specs/001-windows-pager/quickstart.md)
 - [X] T015 [US1] Add an integration test or harness (can be manual-first) at tests/integration/pager_navigation_test.go that validates navigation and search behavior on a Windows runner or simulated TTY
 - [X] T016 [US1] Update specs/001-windows-pager/quickstart.md with explicit interactive validation steps and examples to reproduce acceptance scenarios (file: specs/001-windows-pager/quickstart.md)

---

## Phase 4: User Story 2 - Non-interactive use and automation (Priority: P2)

Goal: Ensure the pager does not interfere with piping, redirection, or automation; provide reliable byte-for-byte output when appropriate.

Independent Test: Run `gyat list | grep something` (or an equivalent pipeline) on Windows and verify the receiving process obtains identical bytes to the unpaged output. Also verify `--no-pager` and GYAT_NO_PAGER disable paging.

### Tests for User Story 2

- [ ] T020 [US2] Ensure writeMaybePagedOutput and internal/pager entry points preserve raw output when pager is bypassed; update cmd/pager.go and internal/pager/pager.go to forward raw bytes to stdout when bypassed (files: cmd/pager.go, internal/pager/pager.go)
- [ ] T021 [US2] Implement binary-vs-text detection helper in internal/pager/output.go and ensure binary output bypasses the pager (file: internal/pager/output.go)
 - [X] T020 [US2] Ensure writeMaybePagedOutput and internal/pager entry points preserve raw output when pager is bypassed; update cmd/pager.go and internal/pager/pager.go to forward raw bytes to stdout when bypassed (files: cmd/pager.go, internal/pager/pager.go)
 - [X] T021 [US2] Implement binary-vs-text detection helper in internal/pager/output.go and ensure binary output bypasses the pager (file: internal/pager/output.go)
- [ ] T022 [P] [US2] Add integration test tests/integration/pager_pipe_test.go that runs a command producing multi-page output, pipes it to a consumer, and asserts the received bytes equal the unpaged output (file: tests/integration/pager_pipe_test.go)
 - [X] T022 [P] [US2] Add integration test tests/integration/pager_pipe_test.go that runs a command producing multi-page output, pipes it to a consumer, and asserts the received bytes equal the unpaged output (file: tests/integration/pager_pipe_test.go)
 - [X] T023 [P] [US2] Add automated tests verifying `--no-pager` and GYAT_NO_PAGER disable the pager across commands (files: cmd/pager_test.go, tests/integration/no_pager_test.go)
 - [X] T022 [P] [US2] Add integration test tests/integration/pager_pipe_test.go that runs a command producing multi-page output, pipes it to a consumer, and asserts the received bytes equal the unpaged output (file: tests/integration/pager_pipe_test.go)
 - [X] T023 [P] [US2] Add automated tests verifying `--no-pager` and GYAT_NO_PAGER disable the pager across commands (files: cmd/pager_test.go, tests/integration/no_pager_test.go)
 - [ ] T024 [P] [US2] Update CLI help text (cmd/status.go) and quickstart docs to clearly document piping/redirection semantics and env/flag overrides (files: cmd/status.go, specs/001-windows-pager/quickstart.md)
  - [X] T024 [P] [US2] Update CLI help text (cmd/status.go) and quickstart docs to clearly document piping/redirection semantics and env/flag overrides (files: cmd/status.go, specs/001-windows-pager/quickstart.md)

---

## Phase 5: Polish & Cross-Cutting Concerns

Purpose: Documentation, tests, release notes, and performance/maintenance improvements that touch multiple stories.

- [ ] T030 [P] Add comprehensive unit and benchmark tests for internal/pager components (files: internal/pager/*_test.go, internal/pager/benchmark_test.go)
- [ ] T031 [P] Update CHANGELOG.md and release notes describing behavior changes (file: CHANGELOG.md)
- [ ] T032 [P] Ensure pager output uses ASCII-safe separators when falling back to `more` on Windows (update code in cmd/pager.go and internal/pager/render.go if needed)
- [ ] T033 [P] Run the quickstart validation steps and record results in specs/001-windows-pager/quickstart.md (file: specs/001-windows-pager/quickstart.md)
 - [X] T030 [P] Add comprehensive unit and benchmark tests for internal/pager components (files: internal/pager/*_test.go, internal/pager/benchmark_test.go)
 - [X] T031 [P] Update CHANGELOG.md and release notes describing behavior changes (file: CHANGELOG.md)
 - [X] T032 [P] Ensure pager output uses ASCII-safe separators when falling back to `more` on Windows (update code in cmd/pager.go and internal/pager/render.go if needed)
 - [X] T033 [P] Run the quickstart validation steps and record results in specs/001-windows-pager/quickstart.md (file: specs/001-windows-pager/quickstart.md)

---

## Dependencies & Execution Order

Phase dependencies:

- Phase 1 (Setup) → Phase 2 (Foundational) → Phase 3 (US1) and Phase 4 (US2) → Phase 5 (Polish)

User story dependencies:

- User Story 1 (P1): Depends on Foundational (T002..T007). Should be implemented first for MVP.
- User Story 2 (P2): Depends on Foundational (T002..T007). Can be implemented in parallel with US1 after foundation is ready.

### Parallel opportunities (examples):

- Create package skeleton (T001) and write package docs (T007) in parallel
- Implement OutputStream helpers (T002), terminal helper (T003) and config (T004) in parallel
- Documentation, changelog and CI workflow tasks (T007, T031, T025) can be run in parallel
- Tests for different stories (T015, T022) can run in parallel where they touch different files

---

## Parallel Execution Example: User Story 1

1. Developer A: T011 implement rendering engine (internal/pager/render.go)
2. Developer B: T012 implement windows input layer (internal/pager/windows.go)
3. Developer C: T013 implement search features (internal/pager/search.go)
4. Developer D: T014 wire pager into CLI (cmd/pager.go, cmd/status.go)

These tasks have clear file boundaries and can be worked on concurrently once foundational tasks complete.

---

## Implementation Strategy

MVP First (User Story 1 Only):

1. Complete Phase 1 (T001)
2. Complete Phase 2 Foundational tasks (T002..T007)
3. Complete Phase 3 User Story 1 tasks (T010..T016) and verify independent test
4. STOP and VALIDATE: Manual/automated interactive tests for US1
5. Proceed to User Story 2 (T020..T025)

Incremental Delivery:

1. Foundation ready → implement US1 (interactive pager) → test and ship as MVP
2. Implement US2 (non-interactive guarantees) → test and ship

---

## Notes

- All tasks follow the checklist format and include explicit file paths.
- IDs are execution-ordered; tasks marked [P] are parallelizable (different files, no cross-task dependency).
- Tests are recommended for acceptance criteria; integration tests are called out where automation is required.
