<!--
Sync Impact Report

Version change: undefined → 1.0.0

Modified principles:
- Added: Workspace-Centric Model
- Added: Thin Git Surface
- Added: Deterministic Execution
- Added: Safe, Non-Destructive Defaults
- Added: CLI-First, Observable, Testable

Added sections:
- Constraints & Technology
- Development Workflow
- Governance: clarified amendment and versioning policy

Removed sections:
- none

Templates requiring review:
- .specify/templates/plan-template.md: ✅ aligned (no mandatory edits required)
- .specify/templates/spec-template.md: ✅ aligned (no mandatory edits required)
- .specify/templates/tasks-template.md: ✅ aligned (no mandatory edits required)
- .specify/templates/commands/*.md: ⚠ none found — verify any custom commands
- README.md: ✅ aligned (describes workspace model & CLI semantics)
- docs/quickstart.md: ⚠ not present (if present, verify CLI/manifest references)

Follow-up TODOs:
- TODO(RATIFICATION_DATE): supply the original ratification date when constitution
  is formally adopted/ratified.
- TODO: designate constitution approvers/maintainers in repository metadata
-->

# gyat Constitution

## Core Principles

### Workspace-Centric Model

- The workspace model is the primary unit of organization. The `.gyat` manifest
  is the single source of truth for tracked repositories, paths, branches, and
  groups. Changes to the manifest or tracked repository layout MUST include a
  documented migration plan and pass the Constitution Check in the implementation
  plan.
- The umbrella repository MUST act as coordination and metadata only. Child
  repositories MUST remain ordinary git repositories and MUST NOT be converted
  into opaque submodules or other forms that reduce visibility or auditability.

Rationale: Preserving child-repository autonomy makes multi-repo changes safer,
auditable, and easier to revert when needed.

### Thin Git Surface

- gyat MUST be a thin wrapper around the git binary. Core VCS semantics and
  behaviors MUST be delegated to git rather than reimplemented.
- Commands MUST preserve familiar git semantics and surface flags that map to
  developer expectations. The tool MUST avoid surprising abstractions.

Rationale: Delegation to git reduces duplication, leverages existing tooling,
and aligns user expectations with familiar behavior.

### Deterministic Execution

- Operations that iterate across multiple targets MUST run in deterministic
  order by default. Determinism is required for reproducible automation and
  scripting.
- Parallel execution is allowed only when explicitly requested (for example
  via `--parallel`). When parallel execution is used, commands MUST clearly
  document ordering and possible interleaving of output.

Rationale: Predictable order and behaviour are essential for reliable
automation, CI, and debugging.

### Safe, Non-Destructive Defaults

- Default command behavior MUST be conservative and non-destructive. Commands
  that remove files, delete working trees, or modify remotes MUST require an
  explicit flag and clear user consent in non-interactive contexts.
- Destructive operations MUST document their scope and provide a safe rollback
  or confirmation path when possible.

Rationale: Protect contributors' local work and prevent accidental mass
deletions or irreversible changes.

### CLI-First, Observable, Testable

- Every user-facing feature MUST expose a clear CLI surface. Outputs intended
  for automation MUST be machine-readable or gated behind flags (for example
  `--json`).
- Output stream conventions MUST be followed: data → stdout; progress, hints,
  warnings, and completion messages → stderr; errors → stderr and non-zero exit
  code. These conventions enable scripting and testing.
- Core commands MUST be covered by tests that exercise success and error
  conditions; command-line behavior MUST be deterministic and testable.

Rationale: A consistent, testable CLI enables automation, integration with CI,
and reliable end-to-end testing.

## Constraints & Technology

- Language & tooling: gyat is implemented in Go and targets Go 1.26+. The CLI
  relies on the system git binary and must not embed a separate git implementation.
- Manifest: `.gyat` is the canonical manifest file. Any changes to its format or
  semantics MUST be backwards-compatible or documented with a migration plan.
- Platforms: gyat targets common development platforms (Linux, macOS, Windows);
  platform-specific behaviour MUST be documented and tested.
- Security: The tool MUST not persist plain-text credentials. Remote URL handling
  MUST treat credentials with caution and defer to git's credential helpers.

## Development Workflow

- Branching/PRs: Changes to gyat or the workspace manifest SHOULD be made in a
  feature branch and submitted as a PR. PR descriptions MUST explain user-facing
  behavior changes and include migration steps for affected workspaces.
- Tests & CI: PRs that change command behavior or manifest handling MUST include
  tests that validate the new behavior and regression tests for previous
  semantics where applicable.
- Constitution Check: Every implementation plan (see
  `.specify/templates/plan-template.md`) MUST include a "Constitution Check"
  section documenting how the proposed change conforms to the principles above.
  This check is a gate that MUST be satisfied before Phase 0 research completes
  and re-checked after Phase 1 design.

## Governance

- Supremacy: This constitution supersedes informal practices. When in conflict,
  follow this document unless a later, ratified amendment states otherwise.
- Amendment procedure:
  - Propose: Amendments are proposed by opening a PR that updates
    `.specify/memory/constitution.md`. The PR MUST include: a concise description
    of the change, rationale, and a migration plan for any affected code or
    workflows.
  - Review: The PR MUST be reviewed and approved by at least two maintainers
    or approvers designated in repository metadata (TODO: designate approvers).
  - Ratification: Merge the PR to the default branch to ratify. The merged file
    MUST set `Last Amended` to the date of merge and update the constitution
    version according to the Versioning Policy below.
- Versioning Policy:
  - MAJOR: Backward-incompatible governance changes, principle removals, or
    redefinitions that materially alter how decisions are made.
  - MINOR: Addition of new principles, new sections, or materially expanded
    guidance that requires new checks or workflows.
  - PATCH: Clarifications, wording fixes, formatting, and non-behavioural
    refinements.
  - The PR that ratifies an amendment MUST include the new semantic version and
    a short rationale for the chosen bump.
- Compliance Reviews: The project SHOULD perform a constitution compliance
  review at least annually and before any major release. Compliance checks MUST
  include verifying that templates (plan/spec/tasks) enforce the Constitution
  Check and that runtime docs reflect current principles.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): initial adoption date
| **Last Amended**: 2026-04-15
