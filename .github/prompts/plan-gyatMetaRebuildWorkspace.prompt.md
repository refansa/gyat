---
description: "Review the current gyat v2 workspace contract, manifest schema, and root discovery"
agent: "plan"
---

# Review Gyat V2 Workspace Contract

Use this prompt to audit the current workspace contract rather than design it
from scratch. `.gyat` manifest v1, root discovery, and `.gitignore`
reconciliation already exist. Focus on whether they are solid enough to carry
into a stable v2 release.

## Current state

- gyat stays in Go.
- `.gyat` is JSON with `version`, `ignore`, and `repos`.
- Each repo currently has `name`, `path`, `url`, optional `branch`, and optional `groups`.
- Workspace root discovery walks upward until it finds `.gyat`.
- The umbrella root remains a normal git repo and tracked child repos remain normal git repos cloned inside it.
- The root `.gitignore` contains a gyat-managed block for tracked repo paths plus manifest ignore patterns.

## Questions to answer

1. Does manifest v1 need any adjustment before release, or should new metadata wait for a later schema version?
2. Are the current path, ignore, and group rules explicit enough for Windows and mixed-path environments?
3. Are there root-discovery or nested-workspace edge cases that still need tightening?
4. Which current behaviors should be documented as part of the contract versus treated as implementation details?
5. What migration or compatibility constraints should limit further schema changes?

## Relevant files

- [internal/manifest/manifest.go](../../internal/manifest/manifest.go)
- [cmd/root.go](../../cmd/root.go)
- [cmd/dir.go](../../cmd/dir.go)
- [cmd/init.go](../../cmd/init.go)
- [internal/workspace/workspace.go](../../internal/workspace/workspace.go)
- [internal/workspace/gitignore.go](../../internal/workspace/gitignore.go)
- [meta/index.js](../../../../mateodelnorte/meta/index.js)
- [meta/lib/findPlugins.js](../../../../mateodelnorte/meta/lib/findPlugins.js)
- [meta/lib/registerPlugin.js](../../../../mateodelnorte/meta/lib/registerPlugin.js)
- [meta/.meta](../../../../mateodelnorte/meta/.meta)

## Deliverables

1. A contract audit with keep, change, and defer recommendations.
2. Any manifest v1 clarifications or blockers that should be resolved before release.
3. A root-discovery and path-handling review, including Windows-specific edge cases.
4. A note on package ownership if runtime boundaries still need cleanup.
5. A list of deferred schema ideas that should not block the first stable v2 release.

## Output format

### Contract summary
Describe the current manifest and workspace contract in practical terms.

### Schema and path risks
Call out concrete issues or ambiguities in manifest fields, path handling, or root discovery.

### Release recommendations
List the changes worth making before release versus after release.

### Deferred ideas
List schema or runtime ideas that should wait for a later version.