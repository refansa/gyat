---
description: "Review the current gyat v2 command surface, selector UX, and legacy migration gaps"
agent: "plan"
---

# Review Gyat V2 Command Surface

Use this prompt to review the current command surface now that the
manifest-based rewrites and shared selector flags are already in place.
Focus on shipped semantics, docs consistency, and remaining migration decisions
rather than redesigning commands from zero.

## Current state

- The CLI remains `gyat <command>` rather than namespaced commands.
- `gyat exec` already exists as the generic multi-repo primitive.
- The command surface is already workspace-based and `.gyat`-driven.
- Shared target flags now exist across workspace-aware commands: `--repo`, `--group`, `--no-root`, `--root-only`, and `--continue-on-error`.
- `init` and `track` reject shared target flags because they are singleton operations.
- `untrack` supports selectors but does not support `--root-only`.

## Commands to review

- `init`
- `track`
- `untrack`
- `rm`
- `list`
- `status`
- `exec`
- `add`
- `commit`
- `pull`
- `push`
- `update`
- `sync`

## Focus areas

1. Validate each command's current semantics rather than redefining the whole surface.
2. Identify help text, README examples, or comments that still describe submodule behavior or outdated root-selection assumptions.
3. Clarify the distinction between path-routing commands (`add`, `rm`, `commit`) and selector-driven commands (`list`, `status`, `exec`, `pull`, `push`, `sync`, `update`, `untrack`).
4. Design the migration or rejection story for legacy `.gitmodules` repositories.
5. Identify any command or flag cleanup that is worth doing before the first stable v2 release.

## Relevant files

- [cmd/init.go](../../cmd/init.go)
- [cmd/track.go](../../cmd/track.go)
- [cmd/untrack.go](../../cmd/untrack.go)
- [cmd/rm.go](../../cmd/rm.go)
- [cmd/list.go](../../cmd/list.go)
- [cmd/status.go](../../cmd/status.go)
- [cmd/exec.go](../../cmd/exec.go)
- [cmd/add.go](../../cmd/add.go)
- [cmd/commit.go](../../cmd/commit.go)
- [cmd/pull.go](../../cmd/pull.go)
- [cmd/push.go](../../cmd/push.go)
- [cmd/update.go](../../cmd/update.go)
- [cmd/sync.go](../../cmd/sync.go)
- [cmd/target_flags.go](../../cmd/target_flags.go)
- [README.md](../../README.md)
- [meta/.meta](../../../../mateodelnorte/meta/.meta)

## Deliverables

1. A command-by-command behavior table for the current v2 surface.
2. A list of docs or help-text inconsistencies versus the implementation.
3. A migration design for repositories that still have `.gitmodules`.
4. A list of semantic or UX risks that still need product decisions before release.
5. A list of commands or flags that should be simplified, renamed, or deferred.

## Output format

### Current behavior table
Describe each command's shipped behavior, inputs, selectors, and outputs.

### Docs and UX gaps
List the biggest mismatches between the implementation and the user-facing docs.

### Migration notes
Explain how current gyat users move from submodules to `.gyat`, or how legacy repos should be handled if migration is deferred.

### Release-facing risks
List the biggest command-level confusion points or rough edges.

### Open questions
List unresolved command semantics that still need a product decision.