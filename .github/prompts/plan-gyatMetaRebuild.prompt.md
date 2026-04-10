---
description: "Review gyat v2 status, the remaining roadmap, and keep the phase prompts aligned"
agent: "plan"
---

# Plan: Gyat V2 Checkpoint

Use this prompt to review the remaining gyat v2 roadmap now that the
manifest-based workspace model and most command rewrites already exist.

## Current baseline

- Keep Go; do not port gyat to Node.
- Breaking change is acceptable, but proposed changes should now be judged against release stability rather than greenfield freedom.
- `.gyat` manifest v1 exists and tracks normal child repos rather than submodules.
- Workspace discovery, target resolution, `.gitignore` reconciliation, and `gyat exec` already exist.
- Most built-in commands already run against the workspace model and share `--repo`, `--group`, `--no-root`, `--root-only`, and `--continue-on-error`.
- The main remaining work is docs and help consistency, migration guidance, deeper tests, and release hardening.

## Focused prompts

- [Workspace contract checkpoint](./plan-gyatMetaRebuildWorkspace.prompt.md)
- [Core orchestration checkpoint](./plan-gyatMetaRebuildCore.prompt.md)
- [Command semantics and migration](./plan-gyatMetaRebuildCommands.prompt.md)
- [Tests, docs, and rollout](./plan-gyatMetaRebuildRollout.prompt.md)

## Current roadmap

### Phase 1 - Workspace contract (mostly complete)
1. Revalidate the manifest schema and root-discovery behavior before v2 stabilizes.
2. Decide whether any runtime or package cleanup is still required before release.

### Phase 2 - Core orchestration (mostly complete)
3. Audit target selection, root-default behavior, and partial-failure semantics for rough edges.
4. Confirm that `exec`, `.gitignore` sync, and workspace loading behavior are documented and well-tested.

### Phase 3 - Command surface and migration (active)
5. Align README, CLI help text, and examples with the shipped manifest-based behavior.
6. Decide how legacy `.gitmodules` repositories should be imported, rejected, or documented.

### Phase 4 - Validation and rollout (active)
7. Fill remaining test gaps for selector-aware commands, migration paths, and Windows behavior.
8. Prepare migration messaging and a release checklist for the first stable v2 cut.

## What to refine

1. Identify remaining release blockers instead of redesigning completed subsystems.
2. Surface inconsistencies between the current implementation and the docs or help text.
3. Call out any schema or UX change that would still justify breaking the current implementation before v2 is declared stable.
4. Break the remaining work into implementation-sized milestones.
5. Push non-blocking extensibility ideas out of the first stable release.

## Output format

### Status summary
Summarize what is already implemented versus what is still open.

### Remaining milestones
Break the outstanding work into implementation-sized milestones with dependencies.

### Release risks
List the main technical, UX, or migration risks still standing between the current tree and a stable v2 release.

### Docs and migration actions
List the documentation and migration tasks that should happen next.

### Deferred items
List ideas worth postponing until after the first stable v2 cut.