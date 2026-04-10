---
description: "Review the current gyat v2 core orchestration, repo targeting, and exec semantics"
agent: "plan"
---

# Review Gyat V2 Core Orchestration

Use this prompt to audit the current orchestration layer now that workspace
loading, target resolution, `.gitignore` reconciliation, and `gyat exec`
already exist.

## Current state

- `internal/git/git.go` remains the only git abstraction.
- `internal/workspace` owns workspace loading, target selection, `.gitignore` reconciliation, and command fan-out primitives.
- Repo targeting already supports root inclusion, root-only selection, explicit repo selectors, and group selectors.
- Workspace-aware commands share `--repo`, `--group`, `--no-root`, `--root-only`, and `--continue-on-error`.
- `gyat exec` is already the generic multi-repo primitive.

## Questions to answer

1. Are package boundaries in `internal/workspace`, `internal/manifest`, and `cmd/` in the right place for a stable v2 release?
2. Are the selector composition rules and root-default behaviors coherent across commands?
3. Is the partial-failure model good enough, or are there commands where `--continue-on-error` should behave differently?
4. Does `gyat exec` have the right output ordering, defaults, and UX for the stable release?
5. What orchestration behaviors still need documentation or deeper tests?

## Relevant files

- [internal/git/git.go](../../internal/git/git.go)
- [internal/workspace/targets.go](../../internal/workspace/targets.go)
- [internal/workspace/runner.go](../../internal/workspace/runner.go)
- [internal/workspace/gitignore.go](../../internal/workspace/gitignore.go)
- [cmd/target_flags.go](../../cmd/target_flags.go)
- [cmd/exec.go](../../cmd/exec.go)
- [cmd/status.go](../../cmd/status.go)
- [cmd/pull.go](../../cmd/pull.go)
- [cmd/push.go](../../cmd/push.go)
- [cmd/testhelper_test.go](../../cmd/testhelper_test.go)
- [meta/index.js](../../../../mateodelnorte/meta/index.js)
- [meta/lib/findPlugins.js](../../../../mateodelnorte/meta/lib/findPlugins.js)

## Deliverables

1. A short architecture audit of the current orchestration layer.
2. A target-selection review with any ambiguities or inconsistencies called out explicitly.
3. A fan-out and failure-model review, including where `--continue-on-error` is or is not appropriate.
4. A concrete `gyat exec` UX review.
5. A focused test plan for any remaining orchestration gaps.

## Output format

### Architecture review
Describe what the current packages own and whether those boundaries still make sense.

### Target selection audit
Explain how selection works today and where it may still confuse users.

### Failure and output model
Review execution ordering, error aggregation, and output behavior.

### `gyat exec` review
Call out any remaining rough edges in `exec` specifically.

### Follow-up tests
List the tests needed to trust the orchestration layer for release.