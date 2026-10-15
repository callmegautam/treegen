package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parse(t *testing.T, input string) []Entry {
	t.Helper()
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return entries
}

func paths(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = filepath.ToSlash(e.Path)
		if e.IsDir {
			out[i] += "/"
		}
	}
	return out
}

func equal(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestParseUnicodeTree(t *testing.T) {
	input := `myapp
├── cmd
│   └── root.go
├── internal
│   ├── a
│   │   └── deep.go
│   └── b.go
└── main.go
`
	equal(t, paths(parse(t, input)), []string{
		"myapp/",
		"myapp/cmd/",
		"myapp/cmd/root.go",
		"myapp/internal/",
		"myapp/internal/a/",
		"myapp/internal/a/deep.go",
		"myapp/internal/b.go",
		"myapp/main.go",
	})
}

func TestSiblingsDoNotNest(t *testing.T) {
	input := `root
├── a.txt
├── b.txt
└── c.txt
`
	equal(t, paths(parse(t, input)), []string{"root/", "root/a.txt", "root/b.txt", "root/c.txt"})
}

func TestParseAsciiTree(t *testing.T) {
	input := "app\n|-- src\n|   `-- main.go\n`-- go.mod\n"

	equal(t, paths(parse(t, input)), []string{"app/", "app/src/", "app/src/main.go", "app/go.mod"})
}

func TestNoRootLine(t *testing.T) {
	input := "├── cmd\n│   └── root.go\n└── main.go\n"

	equal(t, paths(parse(t, input)), []string{"cmd/", "cmd/root.go", "main.go"})
}

func TestDotRootIsNotCreated(t *testing.T) {
	input := ".\n├── cmd\n│   └── root.go\n└── main.go\n"

	equal(t, paths(parse(t, input)), []string{"cmd/", "cmd/root.go", "main.go"})
}

func TestLeafWithoutChildrenIsAFile(t *testing.T) {
	input := ".\n├── cmd\n└── main.go\n"

	equal(t, paths(parse(t, input)), []string{"cmd", "main.go"})
}

func TestTrailingSlashForcesDirectory(t *testing.T) {
	input := "root\n├── empty/\n└── file.txt\n"

	equal(t, paths(parse(t, input)), []string{"root/", "root/empty/", "root/file.txt"})
}

func TestStripsFencesCommentsAndSymlinkTargets(t *testing.T) {
	input := "```text\nroot\n├── main.go   # the entrypoint\n├── link -> ../elsewhere\n└── notes.md // a note\n```\n"

	equal(t, paths(parse(t, input)), []string{"root/", "root/main.go", "root/link", "root/notes.md"})
}

func TestEmptyInput(t *testing.T) {
	if _, err := Parse(strings.NewReader("\n\n```\n```\n")); err == nil {
		t.Error("expected an error for input with no entries")
	}
}

func TestApplyCreatesTree(t *testing.T) {
	root := t.TempDir()
	entries := parse(t, "app\n├── src\n│   └── main.go\n└── go.mod\n")

	plan, err := Apply(root, entries, Options{})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(plan.Created) != 4 {
		t.Errorf("expected 4 created entries, got %d", len(plan.Created))
	}

	if info, err := os.Stat(filepath.Join(root, "app", "src")); err != nil || !info.IsDir() {
		t.Error("app/src should be a directory")
	}
	if info, err := os.Stat(filepath.Join(root, "app", "go.mod")); err != nil || info.IsDir() {
		t.Error("app/go.mod should be an empty file")
	}
}

func TestApplyDryRunTouchesNothing(t *testing.T) {
	root := t.TempDir()
	entries := parse(t, "app\n└── main.go\n")

	if _, err := Apply(root, entries, Options{DryRun: true}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	found, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 0 {
		t.Errorf("dry run wrote to disk: %v", found)
	}
}

func TestApplySkipsExistingFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(root, "app", "main.go")
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := parse(t, "app\n└── main.go\n")
	plan, err := Apply(root, entries, Options{})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(plan.Skipped) != 1 {
		t.Errorf("expected the existing file to be skipped, got %v", plan.Skipped)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "package main" {
		t.Error("an existing file must not be truncated")
	}
}

func TestApplyForceOverwrites(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(root, "app", "main.go")
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := parse(t, "app\n└── main.go\n")
	if _, err := Apply(root, entries, Options{Force: true}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Error("--force should have truncated the file")
	}
}

func TestApplyRefusesToEscapeTheRoot(t *testing.T) {
	root := t.TempDir()

	_, err := Apply(root, []Entry{{Path: filepath.Join("..", "escaped.txt")}}, Options{})
	if err == nil {
		t.Error("expected a refusal to write outside the target directory")
	}
}

func TestIgnoresProseAroundTheTree(t *testing.T) {
	input := "Here is the layout:\nroot\n├── cmd\n│   └── main.go\n└── go.mod\nCopied to clipboard.\nGenerated in 3ms\n"

	equal(t, paths(parse(t, input)), []string{"root/", "root/cmd/", "root/cmd/main.go", "root/go.mod"})
}

func TestStopsAtTheEndOfTheFirstTree(t *testing.T) {
	input := "one\n└── a.txt\ntwo\n└── b.txt\n"

	equal(t, paths(parse(t, input)), []string{"one/", "one/a.txt"})
}
