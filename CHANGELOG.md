# Changelog

## Unreleased

- feat(pager): add cross-platform internal pager with interactive support on Windows
  - internal/pager: rendering, navigation, search
  - internal/pager/windows.go: interactive input/resize handling (Windows)
  - cmd/pager.go: prefer internal pager on Windows and respect GYAT_NO_PAGER

- docs(pager): clarify --no-pager and GYAT_NO_PAGER semantics
  - --no-pager flag and GYAT_NO_PAGER env var force direct streaming to stdout
  - Binary/non-text output is automatically bypassed to avoid mangling raw bytes
  - Added integration test placeholders for manual Windows verification
