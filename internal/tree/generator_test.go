package tree

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func mkTree(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		full := filepath.Join(root, p)
		if strings.HasSuffix(p, "/") {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func generate(t *testing.T, cfg Config) Result {
	t.Helper()
	res, err := Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return res
}

func TestGenerateShape(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root,
		"cmd/root.go",
		"internal/clipboard/clipboard.go",
		"internal/tree/generator.go",
		"internal/tree/ignore.go",
		"go.mod",
		"main.go",
	)

	got := generate(t, Config{Root: root}).Text
	want := strings.Join([]string{
		filepath.Base(root),
		"├── cmd",
		"│   └── root.go",
		"├── internal",
		"│   ├── clipboard",
		"│   │   └── clipboard.go",
		"│   └── tree",
		"│       ├── generator.go",
		"│       └── ignore.go",
		"├── go.mod",
		"└── main.go",
		"",
	}, "\n")

	if got != want {
		t.Errorf("unexpected tree:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "a/b/c/d.txt", "a/b/e.txt", "z.txt", "M.txt", "m.txt", "dir1/", "dir2/x")

	first := generate(t, Config{Root: root}).Text
	for range 20 {
		if got := generate(t, Config{Root: root}).Text; got != first {
			t.Fatalf("output is not deterministic:\n%s\nvs\n%s", first, got)
		}
	}
}

func TestDirectoriesSortBeforeFiles(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "zeta/keep.txt", "alpha.txt")

	lines := strings.Split(strings.TrimSpace(generate(t, Config{Root: root}).Text), "\n")
	if !strings.Contains(lines[1], "zeta") {
		t.Errorf("expected directory first, got %q", lines[1])
	}
}

func TestDepthLimit(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "a/b/c/deep.txt")

	got := generate(t, Config{Root: root, Depth: 2}).Text
	if !strings.Contains(got, "b") {
		t.Error("depth 2 should include the second level")
	}
	if strings.Contains(got, "c") {
		t.Errorf("depth 2 should not descend into the third level:\n%s", got)
	}
}

func TestDepthZeroIsUnlimited(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "a/b/c/d/deep.txt")

	if !strings.Contains(generate(t, Config{Root: root, Depth: 0}).Text, "deep.txt") {
		t.Error("depth 0 should mean unlimited")
	}
}

func TestIgnoreApplies(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "node_modules/pkg/index.js", "src/app.js", "build/out.js", "debug.log")

	m := DefaultMatcher()
	m.Add("build", "*.log")

	got := generate(t, Config{Root: root, Ignore: m}).Text
	for _, gone := range []string{"node_modules", "build", "debug.log"} {
		if strings.Contains(got, gone) {
			t.Errorf("expected %q to be ignored:\n%s", gone, got)
		}
	}
	if !strings.Contains(got, "app.js") {
		t.Error("non-ignored files should survive")
	}
}

func TestEmptyDirectory(t *testing.T) {
	root := t.TempDir()

	got := generate(t, Config{Root: root}).Text
	if got != filepath.Base(root)+"\n" {
		t.Errorf("empty directory should render as just the root, got %q", got)
	}
}

func TestNestedEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "empty/")

	if !strings.Contains(generate(t, Config{Root: root}).Text, "└── empty") {
		t.Error("an empty subdirectory should still be listed")
	}
}

func TestFence(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "main.go")

	got := generate(t, Config{Root: root, Fence: true}).Text
	if !strings.HasPrefix(got, "```text\n") || !strings.HasSuffix(got, "```\n") {
		t.Errorf("expected a fenced block, got:\n%s", got)
	}
}

func TestSymlinkIsNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges on windows")
	}

	root := t.TempDir()
	mkTree(t, root, "real/inside.txt")
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	got := generate(t, Config{Root: root}).Text
	if !strings.Contains(got, "link -> ") {
		t.Errorf("symlink should be rendered with its target:\n%s", got)
	}
	if strings.Count(got, "inside.txt") != 1 {
		t.Errorf("symlinked directory should not be descended into:\n%s", got)
	}
}

func TestSymlinkLoopTerminates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges on windows")
	}

	root := t.TempDir()
	mkTree(t, root, "a/file.txt")
	if err := os.Symlink(root, filepath.Join(root, "a", "loop")); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(generate(t, Config{Root: root}).Text, "loop -> ") {
		t.Error("a symlink loop should render as a link and terminate")
	}
}

func TestUnreadableDirectoryIsReported(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read anything")
	}
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits")
	}

	root := t.TempDir()
	mkTree(t, root, "locked/secret.txt", "open.txt")
	locked := filepath.Join(root, "locked")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })

	res := generate(t, Config{Root: root})
	if !strings.Contains(res.Text, "locked [unreadable]") {
		t.Errorf("unreadable directory should be marked in the output:\n%s", res.Text)
	}
	if len(res.Warnings) != 1 {
		t.Errorf("expected exactly one warning, got %v", res.Warnings)
	}
	if !strings.Contains(res.Text, "open.txt") {
		t.Error("one unreadable directory should not stop the rest of the walk")
	}
}

func TestMissingRootIsAnError(t *testing.T) {
	if _, err := Generate(Config{Root: filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Error("expected an error for a missing root")
	}
}

func TestFileRootIsAnError(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "main.go")

	if _, err := Generate(Config{Root: filepath.Join(root, "main.go")}); err == nil {
		t.Error("expected an error when the root is a file")
	}
}

func TestRelativeRootResolves(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "main.go")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	if !strings.Contains(generate(t, Config{Root: "."}).Text, "main.go") {
		t.Error("a relative root should resolve against the working directory")
	}
}

func TestDefaultIgnoresApplyWhenMatcherIsNil(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, ".git/config", "main.go")

	if strings.Contains(generate(t, Config{Root: root}).Text, ".git") {
		t.Error("a nil matcher should fall back to the default ignore list")
	}
}

func TestGitIgnoreIsRespected(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "src/app.js", "dist/out.js", "debug.log", "keep.log", "sub/nested.log")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n!keep.log\ndist/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := generate(t, Config{Root: root, GitIgnore: true}).Text

	for _, gone := range []string{"dist", "debug.log", "nested.log"} {
		if strings.Contains(got, gone) {
			t.Errorf("expected %q to be gitignored:\n%s", gone, got)
		}
	}
	for _, kept := range []string{"app.js", "keep.log"} {
		if !strings.Contains(got, kept) {
			t.Errorf("expected %q to survive:\n%s", kept, got)
		}
	}
}

func TestNestedGitIgnoreIsRespected(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "a/keep.txt", "a/drop.txt", "b/drop.txt")
	if err := os.WriteFile(filepath.Join(root, "a", ".gitignore"), []byte("drop.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := generate(t, Config{Root: root, GitIgnore: true}).Text

	if strings.Count(got, "drop.txt") != 1 {
		t.Errorf("only b/drop.txt should survive:\n%s", got)
	}
}

func TestGitIgnoreOffByDefault(t *testing.T) {
	root := t.TempDir()
	mkTree(t, root, "dist/out.js")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("dist/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(generate(t, Config{Root: root}).Text, "dist") {
		t.Error("gitignore should only apply when asked for")
	}
}
