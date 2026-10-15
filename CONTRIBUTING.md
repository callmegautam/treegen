# Contributing to treegen

Thanks for taking the time. Bug reports, ideas and patches are all welcome.

## Getting set up

You need Go 1.25.5 or newer. There is nothing else to install.

```bash
git clone https://github.com/callmegautam/treegen.git
cd treegen
go build ./...
go test -race ./...
```

Run your working copy without installing it:

```bash
go run . --no-clipboard ./internal
```

## Before you open a pull request

```bash
gofmt -l .        # must print nothing
go vet ./...
go test -race ./...
```

There is no CI, so these are not run for you. Please run them.

If you touched `install.sh`, run `shellcheck install.sh`. If you touched `install.ps1`, make sure it still parses:

```powershell
$errors = $null
[System.Management.Automation.Language.Parser]::ParseFile("install.ps1", [ref]$null, [ref]$errors)
if ($errors) { $errors }
```

## What a good change looks like

- **Tests.** Every package has a `_test.go` next to it. A fix should come with a case that fails without it. The gitignore tests compare treegen against real `git check-ignore` output, so new pattern handling belongs there.
- **Determinism.** Directories before files, case-insensitive sort. `treegen | treegen new` has to round-trip exactly. If your change can affect ordering, say so in the pull request.
- **No new dependencies** unless there is no reasonable alternative. The binary is static and stays that way.
- **The README is the spec.** If you change a flag, an output format or the parser's tolerance, update the README in the same commit. The example tree in it is real output, so regenerate it with `treegen --ignore '*_test.go'`.

## Project layout

| Path | What lives there |
| --- | --- |
| `cmd/` | Cobra commands: root, `new`, `init` |
| `internal/tree/` | The directory walk, the ignore matcher, the text/json/yaml renderers |
| `internal/scaffold/` | The reverse direction, parsing a tree and creating it on disk |
| `internal/gitignore/` | A from-scratch `.gitignore` engine, differential-tested against git |
| `internal/config/` | `.treegenrc` discovery and parsing |
| `internal/clipboard/` | The one platform-specific corner |

## Commit messages

Short, imperative, lowercase. `fix parser stopping at trailing prose` rather than `Fixed a bug`.

## Reporting a bug

Include the command you ran, what you expected, what happened, and your OS. For parser bugs, paste the exact tree you fed it, inside a code fence so the whitespace survives.
