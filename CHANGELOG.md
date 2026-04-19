# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
-

### Changed
-

### Deprecated
-

### Removed
-

### Fixed
-

### Security
-

## [2.7.0] - 2026-04-19

### Added
- Add tag-release scripts for creating versioned git tags

## [2.6.0] - 2026-04-19

### Added
- Add AGENTS.md and test workspace initialization scripts

### Changed
- Update comments for clarity

## [2.5.0] - 2026-04-15

### Added
- Replace pager with TUI-based viewer

### Changed
- Add replace-pager-with-tui feature spec and checklist

## [2.4.0] - 2026-04-15

### Added
- Add OpenCode/Specify docs and update pager code

### Changed
- Add windows-pager spec & plan

## [2.3.1] - 2026-04-11

### Fixed
- Refine status paging and filtering

## [2.3.0] - 2026-04-10

### Added
- Page status output and align v2 docs

## [2.2.0] - 2026-04-10

### Added
- Add opt-in parallel repo execution

## [2.1.0] - 2026-04-10

### Added
- Unify target flags across commands

## [2.0.1] - 2026-04-10

### Changed
- Drop legacy submodule support

## [2.0.0] - 2026-04-10

### Added
- Add workspace (v2) manifest support

## [0.8.0] - 2026-03-22

### Added
- Add rm command to remove files from working tree and index

### Changed
- Rename previously named remove to untrack

## [0.7.1] - 2026-03-20

### Added
- Add --single-branch flag to track command

### Fixed
- Fix grammatical mistakes in README

## [0.7.0] - 2026-03-19

### Added
- Rename 'remove' command to 'untrack'

## [0.6.0] - 2026-03-19

### Added
- Add status command

## [0.5.1] - 2026-03-19

### Fixed
- Skip pull and push for submodules with local path remotes

## [0.5.0] - 2026-03-19

### Added
- Add pull and push commands

## [0.4.0] - 2026-03-19

### Fixed
- Resolve repo root from binary location; fix porcelain whitespace

## [0.3.0] - 2026-03-19

### Added
- Add commit command for batch-committing across submodules

### Changed
- Add a little bit gag story about the initial idea of the project

## [0.2.3] - 2026-03-19

### Fixed
- Read version from embedded build info when ldflags not used

## [0.2.2] - 2026-03-19

### Added
- Add version variable and --version flag

## [0.2.1] - 2026-03-19

### Added
- Make gyat add a true git add across umbrella and submodules

## [0.2.0] - 2026-03-19

### Added
- Rename `add` to `track`, introduce `gyat add` for staging
- Add go install instructions and fix Go version requirement

## [0.1.0] - 2026-03-19

### Added
- Initial implementation of gyat

[Unreleased]: https://github.com/refansa/gyat/compare/v2.7.0...HEAD
[2.7.0]: https://github.com/refansa/gyat/releases/tag/v2.7.0
[2.6.0]: https://github.com/refansa/gyat/releases/tag/v2.6.0
[2.5.0]: https://github.com/refansa/gyat/releases/tag/v2.5.0
[2.4.0]: https://github.com/refansa/gyat/releases/tag/v2.4.0
[2.3.1]: https://github.com/refansa/gyat/releases/tag/v2.3.1
[2.3.0]: https://github.com/refansa/gyat/releases/tag/v2.3.0
[2.2.0]: https://github.com/refansa/gyat/releases/tag/v2.2.0
[2.1.0]: https://github.com/refansa/gyat/releases/tag/v2.1.0
[2.0.1]: https://github.com/refansa/gyat/releases/tag/v2.0.1
[2.0.0]: https://github.com/refansa/gyat/releases/tag/v2.0.0
[0.8.0]: https://github.com/refansa/gyat/releases/tag/v0.8.0
[0.7.1]: https://github.com/refansa/gyat/releases/tag/v0.7.1
[0.7.0]: https://github.com/refansa/gyat/releases/tag/v0.7.0
[0.6.0]: https://github.com/refansa/gyat/releases/tag/v0.6.0
[0.5.1]: https://github.com/refansa/gyat/releases/tag/v0.5.1
[0.5.0]: https://github.com/refansa/gyat/releases/tag/v0.5.0
[0.4.0]: https://github.com/refansa/gyat/releases/tag/v0.4.0
[0.3.0]: https://github.com/refansa/gyat/releases/tag/v0.3.0
[0.2.3]: https://github.com/refansa/gyat/releases/tag/v0.2.3
[0.2.2]: https://github.com/refansa/gyat/releases/tag/v0.2.2
[0.2.1]: https://github.com/refansa/gyat/releases/tag/v0.2.1
[0.2.0]: https://github.com/refansa/gyat/releases/tag/v0.2.0
[0.1.0]: https://github.com/refansa/gyat/releases/tag/v0.1.0