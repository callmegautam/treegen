# treegen

Generate a clean, markdown-friendly folder tree of a directory and copy it to your clipboard. Then turn one back into real folders and files.

Fast. Deterministic. Cross-platform.

---

## Installation

### Linux and macOS

```bash
curl -fsSL https://gautamsuthar.dev/projects/treegen/install | sh
```

### Windows (PowerShell)

```powershell
irm https://gautamsuthar.dev/projects/treegen/install | iex
```

Both scripts detect your OS and architecture, download the matching release archive, verify it against `checksums.txt`, and drop a single static binary on your PATH. No Go toolchain needed.

That URL serves the shell script by default and the PowerShell one when PowerShell asks for it. To be explicit, or if you would rather read the script before running it, use `/projects/treegen/install.sh` and `/projects/treegen/install.ps1`, which are also the raw files in this repo.

The shell script honours two environment variables:

```bash
curl -fsSL https://gautamsuthar.dev/projects/treegen/install -o install.sh
VERSION=v0.1.0 INSTALL_DIR=~/bin sh install.sh
```

On Windows the equivalents are `TREEGEN_VERSION` and `TREEGEN_INSTALL_DIR`.

### Manual download

Grab the archive for your platform from the [releases page](https://github.com/callmegautam/treegen/releases), extract it, and put the `treegen` binary anywhere on your PATH.

### Using Go

If you already have Go 1.25.5 or newer:

```bash
go install github.com/callmegautam/treegen@latest
```

---

## Usage

Generate a tree for the current directory:

```bash
treegen
```

Or for another directory:

```bash
treegen ./src
```

Limit depth:

```bash
treegen --depth 2
```

Ignore additional folders or files, by name or glob:

```bash
treegen --ignore build,dist,*.log
```

Skip everything your `.gitignore` already skips:

```bash
treegen --gitignore
```

Emit machine-readable output:

```bash
treegen --format json
treegen --format yaml
```

Write to a file instead of the clipboard:

```bash
treegen --output structure.md --fence
```

Pipe just the tree somewhere:

```bash
treegen --quiet --no-clipboard | pbcopy
```

---

## Creating a tree: `treegen new`

The reverse of everything above. Give it a tree and it builds the folders and files.

```bash
treegen new structure.md
pbpaste | treegen new -
treegen new structure.md --into ./scratch --dry-run
```

Paste a structure from a blog post, a README, or a chat with an LLM, and get the skeleton on disk. It reads the same output treegen prints, and also plain ASCII trees using `|--` and `` `-- ``.

What becomes a directory:

- anything ending in `/`
- anything with children under it

Everything else becomes an empty file. Code fences, trailing `# comments`, and ` -> symlink` targets are stripped, so you can paste a block straight out of a document.

Existing files are never touched unless you pass `--force`, and `--dry-run` shows the plan without writing anything.

| Flag | Description |
| --- | --- |
| `-C, --into` | Directory to build the tree in (default `.`) |
| `--dry-run` | Print what would be created, touch nothing |
| `--force` | Overwrite files that already exist |

---

## Config file: `treegen init`

Stop retyping the same flags in the same repo.

```bash
treegen init
```

That writes a commented `.treegenrc`:

```yaml
depth: 0
ignore:
  - dist
  - build
  - "*.log"
gitignore: true
format: text
fence: false
clipboard: true
```

treegen searches upward from the target directory for `.treegenrc`, `.treegenrc.yaml`, `.treegenrc.yml`, `treegen.yaml` or `treegen.yml`, and uses the closest one.

**Flags always win over the file.** The one exception is `ignore`, where the file's list and the flag's list are combined, since both are additions to the built-in defaults. Use `--config <path>` to point at a specific file, or `--no-config` to skip the search entirely.

---

## Flags

| Flag               | Description                                              |
| ------------------ | -------------------------------------------------------- |
| `--depth`          | Limit recursion depth (0 = unlimited)                    |
| `--ignore`         | Comma-separated names or globs to ignore, added to the defaults |
| `--gitignore`      | Also skip anything the repository's `.gitignore` files skip |
| `-f, --format`     | Output format: `text`, `json` or `yaml`                  |
| `-o, --output`     | Write the tree to a file instead of the clipboard        |
| `--fence`          | Wrap text output in a markdown code fence                |
| `--no-clipboard`   | Print the tree without copying it                        |
| `-q, --quiet`      | Print only the tree, no status lines                     |
| `--config`         | Use a specific config file                               |
| `--no-config`      | Ignore any config file                                   |
| `-v, --version`    | Print the version                                        |

---

## Default Ignored

The following are ignored automatically:

- `.git`
- `node_modules`
- `.DS_Store`
- `.idea`
- `.vscode`

`--ignore` extends this list rather than replacing it. Patterns are matched against each entry's name at every level, and support the usual glob syntax (`*.log`, `tmp?`).

`--gitignore` is separate and additive. It reads the `.gitignore` in the target directory and in every subdirectory it walks, and applies real git semantics: negation with `!`, directory-only patterns with a trailing `/`, anchoring with a leading `/`, `**`, and last-match-wins. The implementation is differential-tested against `git check-ignore`.

---

## Example Output

Run inside this repository:

```console
$ treegen --ignore '*_test.go'
treegen
├── .github
│   ├── ISSUE_TEMPLATE
│   │   ├── bug_report.yml
│   │   ├── config.yml
│   │   └── feature_request.yml
│   └── PULL_REQUEST_TEMPLATE.md
├── cmd
│   ├── init.go
│   ├── new.go
│   └── root.go
├── internal
│   ├── clipboard
│   │   └── clipboard.go
│   ├── config
│   │   └── config.go
│   ├── gitignore
│   │   └── gitignore.go
│   ├── scaffold
│   │   └── scaffold.go
│   └── tree
│       ├── format.go
│       ├── generator.go
│       └── ignore.go
├── .gitignore
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── install.ps1
├── install.sh
├── LICENSE
├── main.go
├── README.md
├── release.sh
└── SECURITY.md
Copied to clipboard.
Generated in 3ms
```

Directories are listed before files, and entries are sorted case-insensitively, so the same directory always produces the same tree. That determinism is what makes `treegen | treegen new` round-trip exactly.

With `--format json` the same walk comes out as structured data:

```json
{
  "name": "treegen",
  "type": "directory",
  "children": [
    { "name": "main.go", "type": "file" }
  ]
}
```

Nodes carry `type` (`directory`, `file` or `symlink`), a `target` for symlinks, and an `error` when a directory could not be read.

---

## Behaviour notes

- Symlinks are listed with their target (`link -> ../real`) and are never followed, so symlink loops terminate.
- A directory that cannot be read is marked `[unreadable]` in the tree and reported on stderr; the rest of the walk continues.
- If no clipboard is available (headless, ssh, CI), treegen prints a warning on stderr and still exits 0. Use `--no-clipboard` to skip the copy entirely.
- The binary is fully static with no runtime dependencies. The clipboard is the one exception on Linux, where it shells out to `xclip`, `xsel`, or `wl-copy`. Windows uses the native clipboard API and macOS uses `pbcopy`, both built in.

---

## Development

```bash
go build ./...
go test -race ./...
```

Releases are cut from a working copy, there is no CI:

```bash
./release.sh v0.1.0
./release.sh v0.1.0 --publish
```

The first form cross-compiles all six targets into `dist/` and writes `checksums.txt`. The second also creates the GitHub release with `gh` and uploads the archives. The archive names are what the install scripts download, so they have to stay in step.

---

## Contributing

Bug reports and pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md)
for how to get set up and what to run before you open one, and
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for the ground rules.

Found a security issue? Please report it privately, see [SECURITY.md](SECURITY.md).

Release history lives in [CHANGELOG.md](CHANGELOG.md).

---

## License

MIT, see [LICENSE](LICENSE).
