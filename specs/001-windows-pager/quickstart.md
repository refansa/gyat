# Quickstart: windows-pager

How to exercise the Windows pager feature locally:

1. Build or run the gyat CLI on Windows.
2. Ensure you are in a workspace with enough tracked repos to produce multi-page output
   (for example, a `.gyat` manifest with many repos).
3. Run a command that produces long output, e.g., `gyat list`.
4. Interact with pager: use space to page forward, `b` for page back, `/` to search,
   `n`/`N` to navigate matches, and `q` to quit.
5. To verify non-interactive behavior, pipe the command: `gyat list | grep something`
    and confirm the receiving program gets unmodified output.

Notes:
- If your terminal does not support resize events, the pager will use a fixed
  viewport based on the initial terminal size.
- Use `--no-pager` or set `GYAT_NO_PAGER=1` to force streaming output to stdout.
 - Binary or non-text output is automatically bypassed and written directly to
   stdout to avoid mangling raw bytes. The pager will not be invoked for such
   output.

Interactive validation steps (Windows)

1. Build the gyat CLI on Windows and run inside a terminal (Windows Terminal, PowerShell).
2. Ensure your workspace produces multi-page output (e.g., many tracked repos) and run:

   gyat status

3. Interact with the pager:
   - Space: page forward
   - b: page back
   - j / k: line down / up
   - /<query><Enter>: search for a term, the pager will jump to the first match
   - n / N: jump to next / previous match
   - q: quit the pager and return to the shell

4. Verify that when `--no-pager` is supplied or `GYAT_NO_PAGER=1` is set the output is streamed directly to stdout (no interactive behavior).

Manual integration test note:
 - Automated interactive tests require a pseudo-tty; the integration test harness is documented in tests/integration/pager_navigation_test.go and is intended for manual execution on a Windows runner.

Validation Results (automated checks)

1. Unit tests: internal pager unit tests and cmd pager tests are present. Benchmarks were added to internal/pager/benchmark_test.go to exercise Render and detection heuristics.
2. Integration placeholders: interactive navigation and no-pager integration tests are documented as manual steps and have placeholders in tests/integration/ to support running on Windows runners.

Notes on manual validation:
- Interactive navigation and resize behavior must be validated manually on a Windows terminal (Windows Terminal or PowerShell) as described in the steps above. Automated interactive verification is out-of-scope for CI without a dedicated TTY harness.
