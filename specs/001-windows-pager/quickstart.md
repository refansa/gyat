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
