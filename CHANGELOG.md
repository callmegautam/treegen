# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - unreleased

First release.

### Added

- `treegen` generates a markdown-friendly tree of a directory and copies it to
  the clipboard.
- `--depth` to limit recursion, `--ignore` for extra names and globs on top of
  the built-in defaults.
- `--gitignore` applies real git ignore semantics, including negation, anchoring,
  directory-only patterns, `**`, and last-match-wins. Differential-tested against
  `git check-ignore`.
- `--format json` and `--format yaml` for machine-readable output.
- `--output` to write to a file, `--fence` to wrap in a markdown code fence,
  `--no-clipboard` and `--quiet` for piping.
- `treegen new` turns a tree back into real folders and files, reading treegen's
  own output as well as plain ASCII trees. Supports `--into`, `--dry-run` and
  `--force`.
- `treegen init` writes a commented `.treegenrc`. Config files are discovered by
  searching upward; flags win, except `ignore`, which merges.
- Install scripts for Unix and Windows that verify downloads against
  `checksums.txt`.
- `release.sh` cross-compiles all six targets and optionally publishes the
  GitHub release.

[0.1.0]: https://github.com/callmegautam/treegen/releases/tag/v0.1.0
