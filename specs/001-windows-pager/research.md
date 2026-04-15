# Research: windows-pager

Date: 2026-04-15

This document captures research tasks and decisions required to adapt the Unix-like
pager behavior to the Windows implementation.

## Decisions and Rationale

1. Decision: Use existing Unix-like pager implementation as reference and extract
   cross-platform pager logic into a shared module.

   Rationale: The Unix implementation already provides the desired interaction
   model and test coverage. Reusing logic reduces duplication and keeps behaviour
   consistent across platforms.

2. Decision: Implement a Windows adaptation layer that uses console APIs for
   key input and resizing where available and falls back to a best-effort mode
   on limited terminals.

   Rationale: Windows console APIs differ from POSIX terminals; a thin layer
   keeps platform-specific code isolated.

3. Decision: Pager invoked only on TTY and when not piping/redirection; offer
   `--no-pager` and `GYAT_NO_PAGER` env var to force non-interactive output.

   Rationale: Ensures automation compatibility and explicit override for users.

## Research Tasks

- Research 1: Audit the repository to find the Unix-like pager implementation and
  identify reusable components (API surface, internal helpers).
- Research 2: Evaluate terminal libraries available in the project's language that
  support Windows consoles and POSIX terminals; prefer libraries already in use.
- Research 3: Identify behavior differences in Windows consoles (key codes,
  resize events, input buffering) and plan adaptation strategies.

## Findings (initial)

- The Unix-like pager is present in the codebase (developer observation / user note).
  Locate its module during implementation to extract behavior.
- Windows terminals vary; Windows Terminal and modern PowerShell support ANSI
  and many terminal features. Older cmd.exe may be limited; document differences.
