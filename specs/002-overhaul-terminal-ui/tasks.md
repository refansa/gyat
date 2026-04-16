---

description: "Task list for feature Overhaul Terminal UI"
---

# Tasks: Overhaul Terminal UI

**Input**: Design documents from `C:\Users\user\workspace\github.com\refansa\gyat\specs\002-overhaul-terminal-ui/`

**Prerequisites**: `plan.md` (required), `spec.md` (required for user stories), `research.md`, `data-model.md`, `contracts/` (if present)

## Phase 1: Setup (Project initialization)

**Purpose**: Add project scaffolding, dependencies and test harness required before any story work begins.

- [X] T001 [P] Create package skeleton for the interactive UI under C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui: add `doc.go`, `model.go`, `list.go`, `tabs.go` (empty skeletons with package comments)
- [X] T002 [P] Ensure Bubble Tea stack dependencies are declared in `C:\Users\user\workspace\github.com\refansa\gyat\go.mod` (run `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles github.com/charmbracelet/lipgloss` and commit updated `go.mod`/`go.sum`)
- [X] T003 Add a persistent `--no-ui` opt-out flag and wiring in `C:\Users\user\workspace\github.com\refansa\gyat\cmd\root.go` and ensure `C:\Users\user\workspace\github.com\refansa\gyat\cmd\list.go` and `C:\Users\user\workspace\github.com\refansa\gyat\cmd\status.go` respect the flag
- [X] T004 [P] Add a TTY-driven smoke test harness at `C:\Users\user\workspace\github.com\refansa\gyat\tests\integration\tui_smoke_test.go` that can build the binary and exercise `gyat list` in a pseudo-tty for manual/CI smoke validation

---

## Phase 2: Foundational (Blocking prerequisites)

**Purpose**: Implement core models, deterministic ordering and non-blocking data retrieval used by all stories.

**⚠️ CRITICAL**: No user story work should begin until these tasks are complete.

- [X] T005 Create data model structs to match `data-model.md` in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\model\repository.go`:
  - `RepositoryEntry` (id, display_name, path, current_branch, summary_state, metadata)
  - `RepositoryStatusView` (repo_id, tabs, active_tab)
  - `StatusTab` (id, title, content)
- [X] T006 Add deterministic ordering helper to `C:\Users\user\workspace\github.com\refansa\gyat\internal\manifest\manifest.go` (exported function `OrderRepositoriesByManifest(manifest *manifest.Manifest, repos []workspace.Repository) []workspace.Repository`) per research decision to respect `.gyat` manifest grouping
- [X] T007 [P] Implement an asynchronous repository summary fetcher in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\data\fetcher.go` that:
  - accepts a list of repository paths, fetches git-derived summary data concurrently, and returns `[]RepositoryEntry`
  - never blocks the Bubble Tea update loop (use goroutines + channels or context-aware workers)
- [X] T008 Create the Bubble Tea model skeleton that composes the list and tabs views in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\model.go` (implement Init, Update, View stubs and message types)
- [X] T009 Add non-UI fallback helpers in `C:\Users\user\workspace\github.com\refansa\gyat\cmd\list.go` and `C:\Users\user\workspace\github.com\refansa\gyat\cmd\status.go` so commands run the legacy plain-text output when `--no-ui` is set or stdout is not a TTY

---

## Phase 3: User Story 1 - Browse repositories interactively (Priority: P1) 🎯 MVP

**Goal**: Replace pager-style repository listing with an interactive list that supports keyboard navigation and keeps the current selection visible.

**Independent Test**: Run the repository listing experience with multiple repositories present and verify the user can move through a visible list, keep track of the current selection, and review repository information without leaving the screen.

### Tests (write first)

- [X] T010 [P] [US1] Add unit tests for selection, paging and visible-selection guarantees in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\model_test.go`

### Implementation

- [X] T011 [US1] Implement the interactive list view using Bubbles in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\list.go` (render grouped repositories, current selection highlight, scrolling behavior)
- [X] T012 [US1] Wire the list UI into the `gyat list` command in `C:\Users\user\workspace\github.com\refansa\gyat\cmd\list.go` so the UI launches when `--no-ui` is false and stdout is a TTY (fall back to legacy output otherwise)
- [X] T013 [US1] Implement keyboard navigation and keybindings (up/down, page up/down, home/end) for the list in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\keys.go` or inside `model.go`
- [X] T014 [P] [US1] Add an integration test `C:\Users\user\workspace\github.com\refansa\gyat\tests\integration\tui_list_test.go` that uses the TTY smoke harness to assert the list opens and selection can move

**Checkpoint**: After T014, the interactive repository list must be usable and independently testable

---

## Phase 4: User Story 2 - Inspect per-repository status with tabs (Priority: P2)

**Goal**: Provide tabbed status views for the selected repository and allow switching tabs without leaving the interface.

**Independent Test**: Open the repository status experience, select a repository, switch between available tabs, and confirm the content updates for the selected repository while preserving overall interface context.

### Tests (write first)

- [X] T015 [P] [US2] Add unit tests for tab switching and active-tab stability in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\tabs_test.go`

### Implementation

- [X] T016 [US2] Implement the tabs/status pane in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\tabs.go` (render active tab, update content when selection changes)
- [X] T017 [US2] Integrate the status TUI with the `gyat status` command in `C:\Users\user\workspace\github.com\refansa\gyat\cmd\status.go` so the tabbed view opens for the selected repository when UI is enabled
- [X] T018 [US2] Add an integration test `C:\Users\user\workspace\github.com\refansa\gyat\tests\integration\tui_status_test.go` verifying tab switching and content updates

**Checkpoint**: After T018 the status tabs should be functional and independently testable

---

## Phase 5: User Story 3 - Use the same interaction model across commands (Priority: P3)

**Goal**: Ensure `list` and `status` share the same navigation and keybinding model so users don't need to relearn interaction patterns.

**Independent Test**: Launch `gyat list` and `gyat status` (UI enabled) and verify the same navigation keys produce the same behaviors in both contexts.

### Implementation

- [X] T019 [P] [US3] Refactor shared UI primitives into `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\components\` (create `list.go`, `tabs.go`) and update `internal/ui/tui` to reuse them
- [X] T020 [US3] Centralize keybindings in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\keys.go` and update `C:\Users\user\workspace\github.com\refansa\gyat\cmd\list.go` and `C:\Users\user\workspace\github.com\refansa\gyat\cmd\status.go` to use the shared keymap
- [X] T021 [US3] Add an end-to-end consistency integration test `C:\Users\user\workspace\github.com\refansa\gyat\tests\integration\tui_consistency_test.go` that verifies equivalent navigation across `list` and `status`

**Checkpoint**: After T021 both commands should feel consistent and reuse the same primitives

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, performance, and final polishing that affect multiple stories

- [X] T022 [P] Update `C:\Users\user\workspace\github.com\refansa\gyat\specs\002-overhaul-terminal-ui\quickstart.md` with the new TUI usage, `--no-ui` examples and keybindings summary
- [X] T023 [P] Implement an interactive help/cheatsheet in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\help.go` and ensure `?` / `h` toggles it
- [X] T024 [P] Add benchmark tests in `C:\Users\user\workspace\github.com\refansa\gyat\internal\ui\tui\benchmark_test.go` to measure navigation latency against a 200-repo fixture and target <50ms navigation latency
- [ ] T025 [P] Finalize documentation and plan updates in `C:\Users\user\workspace\github.com\refansa\gyat\specs\002-overhaul-terminal-ui\plan.md` and `C:\Users\user\workspace\github.com\refansa\gyat\specs\002-overhaul-terminal-ui\quickstart.md` and add release notes in `C:\Users\user\workspace\github.com\refansa\gyat\specs\002-overhaul-terminal-ui\release-notes.md`

---

## Dependencies & Execution Order

- Phase 1 (Setup) tasks: T001 → T002 → T003 → T004 (can start immediately; T002/T004 parallelizable)
- Phase 2 (Foundational) tasks: T005..T009 — BLOCK all user stories; must complete before Phase 3
- User Stories (Phase 3+): Start after Foundational completion. Recommended order: US1 (T010..T014) → US2 (T015..T018) → US3 (T019..T021), but stories can be implemented in parallel if team capacity allows
- Polish (Final Phase): T022..T025 — after desired stories are implemented

### User Story Dependencies

- **US1**: No dependencies on other stories (depends on Foundational phase)
- **US2**: Depends on Foundational; integrates with US1 components but must be independently testable
- **US3**: Depends on Foundational; refactors shared components used by US1 and US2

---

## Parallel execution examples (per story)

- US1 parallel example (after Foundational):
  - Run T011 (implement list view), T013 (keybindings) and T010 (unit tests) in parallel if different engineers are available
- US2 parallel example:
  - Run T016 (tabs implementation) and T015 (unit tests) concurrently
- Cross-story parallelism:
  - T019 (refactor shared components) can proceed in parallel with T021 (end-to-end tests) if isolation is preserved

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (Setup)
2. Complete Phase 2 (Foundational)
3. Implement Phase 3 (US1) and verify independently (T010..T014)
4. STOP and VALIDATE: run the TTY smoke harness and manual validation

### Incremental Delivery

1. Foundation ready → implement US1 → validate and release MVP
2. Add US2 → validate and release
3. Add US3 → validate and release

---

## Validation Checklist

- Total tasks: 25
- Task counts by tranche:
  - Setup (Phase 1): 4
  - Foundational (Phase 2): 5
  - User Story 1 (US1): 5
  - User Story 2 (US2): 4
  - User Story 3 (US3): 3
  - Polish & Cross-Cutting: 4

- Parallel opportunities identified (tasks marked [P]): T001, T002, T004, T007, T010, T014, T015, T019, T022, T023, T024, T025

- Independent test criteria per story (summary):
  - US1: Browse repository list with multiple repos, verify navigation, visible selection and repository detail inspection without leaving screen
  - US2: Select a repository and switch tabs; active tab and content update correctly; selection changes update tabs
  - US3: `list` and `status` produce the same navigation and selection behaviors across commands

- Suggested MVP scope: Complete Setup + Foundational + User Story 1 (T001..T014)

- Format validation: All tasks follow the required checklist format `- [ ] TNNN [P?] [US?] Description with file path` (IDs sequential, story labels present only for user story tasks)
