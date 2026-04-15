Pager Contract (internal/pager)

Purpose

This contract documents the public behavior and interface expectations for the internal pager package used by the gyat CLI.

Public API (stable surface)

- NewPager(w io.Writer) *Pager
  - Creates a pager bound to a writer (stdout). The writer may be an *os.File or a test writer.

- (*Pager).Render(content []byte) (int, error)
  - Initializes pager state and writes the first page to the writer. Returns bytes written. Must not alter the content bytes returned to pipes when stdout is not a TTY.

- (*Pager).HandleKey(r rune) (int, error)
  - Apply a single key action and render resulting visible page. Keys supported: 'q',' ','b','j','k','n','N','/' (search acts via Read input in interactive loop). Unknown keys are no-ops.

- (*Pager).Search(query string) int
  - Record substring matches (case-sensitive for MVP) and return match count.

- (*Pager).Seek(index int) (int, error)
  - Move viewport to the given match index and render page.

- RunInteractiveSession(p *Pager, in *os.File, out *os.File) error
  - Start interactive loop; returns when session ends.

Behavioral guarantees

- Non-interactive contexts (stdout not a TTY) must not start interactive sessions and must emit raw content unchanged.
- When RunInteractiveSession is used, it must put the input terminal into raw mode and restore it on exit.
- Exit codes and stdout stream behavior of the invoking command must be preserved (the pager must not swallow or modify trailing newlines or change exit codes).

Backward compatibility rules

- Any API additions must be additive; do not remove existing exported functions.
- Keep the Pager type and method signatures stable unless an explicit migration plan is approved.

Usage examples

1) Non-interactive (pipe): write raw bytes to stdout (no pager invoked)

2) Interactive: create a Pager with NewPager(os.Stdout), call Render, then RunInteractiveSession
