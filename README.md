# treegen

Generate a clean, markdown-friendly folder tree of your current directory and copy it to your clipboard.

Fast. Deterministic. Cross-platform.

---

## Installation

### Using Go

```bash
go install github.com/YOUR_USERNAME/treegen@latest
```

Make sure your `$GOBIN` or `$GOPATH/bin` is in your PATH.

### Using Prebuilt Binary

Download the appropriate binary for your OS from the **Releases** section and place it in a directory included in your PATH.

---

## Usage

Generate tree for the current directory:

```bash
treegen
```

Limit depth:

```bash
treegen --depth 2
```

Ignore additional folders or files:

```bash
treegen --ignore build,dist
```

---

## Flags

| Flag       | Description                             |
| ---------- | --------------------------------------- |
| `--depth`  | Limit recursion depth (0 = unlimited)   |
| `--ignore` | Comma-separated folders/files to ignore |

---

## Default Ignored

The following are ignored automatically:

- `.git`
- `node_modules`
- `.DS_Store`
- `.idea`
- `.vscode`

---

## Example Output

```text
.
├── cmd
│   └── root.go
├── internal
│   ├── clipboard
│   │   └── clipboard.go
│   └── tree
│       ├── generator.go
│       └── ignore.go
├── go.mod
└── main.go
```

The generated output is automatically copied to your clipboard.
