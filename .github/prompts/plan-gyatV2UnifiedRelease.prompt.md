# Plan: Gyat V2 Unified Release Plan

Reconcile the existing planning prompts against the current tree, treat the manifest/workspace rewrite as already shipped, and focus the remaining roadmap on release hardening. The unified plan below marks completed foundations, isolates the still-open release work, and avoids re-planning subsystems that are already implemented.

## Status Summary

- Done: Workspace contract foundations are in place. `.gyat` manifest v1, validation, root discovery, target resolution, `.gitignore` reconciliation, and the workspace runner exist and are tested.
- Done: The command surface has already been rewritten around the workspace model. `init`, `track`, `list`, `status`, `exec`, `add`, `commit`, `pull`, `push`, `update`, `sync`, `rm`, and `untrack` exist with shared selector behavior where appropriate.
- Done: README is largely v2-oriented and command coverage is broad. Command tests, internal package tests, and at least one end-to-end workflow test already exist.
- Open: A few implementation seams still need alignment before calling v2 stable, especially local-path clone handling in `track` and the need to document `--parallel` on a command-by-command basis.
- Open: The checked-in `.github/copilot-instructions.md` still documents the older submodule-era model and should be updated or explicitly scoped so it stops fighting the current implementation.

## Steps

### 1. Phase 1 — Lock the current v2 contract as the release baseline

Document `.gyat` manifest v1, workspace root discovery, target-selection rules, and `.gitignore` managed-block behavior as stable release contract. Treat schema redesign as out of scope unless a concrete release blocker is found. This phase revalidates what is already implemented rather than changing it.

### 2. Phase 2 — Resolve the remaining command-surface inconsistencies

Audit workspace-aware commands for flag and behavior consistency, with special attention to `--parallel`, root defaults, selector interactions, and help text. Document `--parallel` as a command-specific decision rather than forcing one universal rule across `commit`, `pull`, `push`, `update`, and `exec`. This step depends on Phase 1 because the command semantics should align to the frozen contract.

### 3. Phase 3 — Close the local-path gap

Align local-path handling so `track` and other local-remote flows consistently apply the intended `protocol.file.allow=always` behavior where required. Keep this phase tightly scoped to shipped workspace behavior rather than reopening broader migration or compatibility work.

### 4. Phase 4 — Finish release-facing documentation

Update README, CLI help text, examples, and checked-in agent instructions so they consistently describe the shipped workspace model, selector defaults, and any approved command-specific parallel behavior. This can begin after the Phase 2 decisions are made and should be completed after Phase 3 finalizes the local-path behavior.

### 5. Phase 5 — Fill the targeted validation gaps

Extend tests where current coverage is thin: selector consistency across commands, local-path clone/update behavior, and Windows-specific cases such as drive-letter paths, UNC paths, and cross-platform path normalization. Add CI coverage for Windows if release scope allows. This can run in parallel with Phase 4 once Phases 2 and 3 settle the intended behavior.

### 6. Phase 6 — Prepare the release gate

Create a concrete v2 release checklist covering test commands, supported platforms, docs sign-off, and tag/release steps. Treat stale top-level instructions and missing Windows validation as blockers for the first stable v2 cut.

## Relevant files

- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/internal/manifest/manifest.go` — existing manifest schema, normalization, and validation to keep as the v2 contract baseline.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/internal/workspace/workspace.go` — current upward root discovery and workspace loading.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/internal/workspace/gitignore.go` — gyat-managed `.gitignore` reconciliation that is already part of the shipped model.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/internal/workspace/targets.go` — selector composition rules and root/default targeting behavior.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/internal/workspace/runner.go` — ordered serial/parallel fan-out behavior to reuse when deciding command consistency.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/target_flags.go` — shared selector and run-option flags, including `--parallel` plumbing.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/track.go` — local-path clone flow and current missing alignment with `protocol.file.allow=always`.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/update.go` — current local-path-safe update behavior to use as the reference implementation.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/commit.go` — workspace-aware command that currently does not bind `--parallel`.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/pull.go` — workspace-aware command needing final decision on serial vs parallel behavior.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/push.go` — workspace-aware command needing final decision on serial vs parallel behavior.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/exec.go` — generic multi-repo primitive whose parallel behavior should be made explicit.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/README.md` — mostly updated v2 docs, but still missing release-readiness and command-consistency notes.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/.github/copilot-instructions.md` — stale submodule-era instructions that should be aligned or quarantined.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/integration_test.go` — current end-to-end coverage baseline to extend.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/testhelper_test.go` — reusable repo fixtures and cross-platform path helpers.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/internal/workspace/targets_test.go` — selector behavior tests to expand.
- `c:/Users/thumbleweed/workspace/github.com/refansa/gyat/cmd/track_test.go` — local-path and Windows-path expectations.

## Verification

1. Recheck the finished contract against the implementation by reviewing manifest validation, root discovery, selector resolution, and `.gitignore` reconciliation in the files above.
2. Run the existing command and internal suites: `go test ./...`.
3. Add or update targeted tests for any chosen command-semantics changes, then rerun focused suites such as `go test ./cmd/... -run TestRunTrack`, `go test ./cmd/... -run TestRunStatus`, `go test ./cmd/... -run TestRunAdd`, and `go test ./internal/workspace/...`.
4. If Windows is part of the stable-release promise, validate the suite on Windows or add a CI matrix that proves it.
5. Manually verify docs/help alignment by comparing README examples and command help against actual runtime behavior for `track`, `exec`, `pull`, `push`, `update`, `status`, and `untrack`.

## Decisions

- Included scope: release-hardening work for the already-shipped v2 workspace model.
- Excluded scope: new manifest schema fields, a redesign of the command surface, and post-v2 extensibility ideas unless they fix a concrete release blocker.
- Recommended baseline: treat the prompt set in `.github/prompts` as the accurate roadmap source and treat `.github/copilot-instructions.md` as stale until updated.
- Recommended product decision: keep `--parallel` command-specific and document each command's behavior explicitly before expanding tests or rewriting help text.

## Further Considerations

1. Parallel policy should stay explicit and command-specific. Recommendation: document it per command instead of promoting it to a universal workspace flag.
2. Windows validation can be staged. Recommendation: add explicit Windows-path tests immediately, then decide whether Windows CI is required before tagging the first stable release.
3. Stale checked-in instructions are a release risk for future contributors. Recommendation: update or replace `.github/copilot-instructions.md` during the docs phase instead of leaving it for after release.
