package gitignore

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func stackFor(t *testing.T, root string, dirs ...string) *Stack {
	t.Helper()
	st := NewStack()
	if err := st.PushDir(root); err != nil {
		t.Fatal(err)
	}
	for _, d := range dirs {
		if err := st.PushDir(filepath.Join(root, d)); err != nil {
			t.Fatal(err)
		}
	}
	return st
}

func TestBasicPatterns(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".gitignore"), strings.Join([]string{
		"# a comment",
		"",
		"*.log",
		"build/",
		"/only-at-root.txt",
		"docs/**/draft.md",
		"temp?",
	}, "\n"))

	st := stackFor(t, root)

	cases := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{"debug.log", false, true},
		{"nested/debug.log", false, true},
		{"notes.md", false, false},
		{"build", true, true},
		{"build/out.js", false, true},
		{"src/build", true, true},
		{"only-at-root.txt", false, true},
		{"nested/only-at-root.txt", false, false},
		{"docs/a/b/draft.md", false, true},
		{"docs/draft.md", false, true},
		{"temp1", false, true},
		{"temp", false, false},
	}

	for _, c := range cases {
		got := st.Ignored(filepath.Join(root, c.path), c.isDir)
		if got != c.want {
			t.Errorf("Ignored(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestNegation(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".gitignore"), "*.log\n!keep.log\n")

	st := stackFor(t, root)

	if !st.Ignored(filepath.Join(root, "debug.log"), false) {
		t.Error("debug.log should be ignored")
	}
	if st.Ignored(filepath.Join(root, "keep.log"), false) {
		t.Error("keep.log should be re-included by the negation")
	}
}

func TestNegationCannotEscapeAnIgnoredDirectory(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".gitignore"), "build/\n!build/keep.txt\n")

	st := stackFor(t, root)

	if !st.Ignored(filepath.Join(root, "build", "keep.txt"), false) {
		t.Error("git does not re-include files inside an ignored directory")
	}
}

func TestNestedGitignoreWins(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".gitignore"), "*.log\n")
	write(t, filepath.Join(root, "sub", ".gitignore"), "!important.log\n")

	st := stackFor(t, root, "sub")

	if st.Ignored(filepath.Join(root, "sub", "important.log"), false) {
		t.Error("the deeper .gitignore should win")
	}
	if !st.Ignored(filepath.Join(root, "sub", "other.log"), false) {
		t.Error("the root pattern still applies where the nested file is silent")
	}
}

func TestEmptySet(t *testing.T) {
	st := NewStack()
	if st.Ignored("/anything", false) {
		t.Error("an empty stack should ignore nothing")
	}
	if err := st.PushDir(t.TempDir()); err != nil {
		t.Errorf("a missing .gitignore is not an error: %v", err)
	}
}

func gitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func TestMatchesGitCheckIgnore(t *testing.T) {
	if !gitAvailable() {
		t.Skip("git is not installed")
	}

	root := t.TempDir()

	rules := strings.Join([]string{
		"*.log",
		"!keep.log",
		"build/",
		"/root-only.txt",
		"node_modules",
		"docs/**/draft.md",
		"**/generated",
		"temp?",
		"*.tmp",
		"!src/*.tmp",
		"a/b/c.txt",
	}, "\n")
	write(t, filepath.Join(root, ".gitignore"), rules+"\n")
	write(t, filepath.Join(root, "sub", ".gitignore"), "local.txt\n!../unreachable\n")

	paths := []struct {
		path  string
		isDir bool
	}{
		{"debug.log", false},
		{"keep.log", false},
		{"sub/debug.log", false},
		{"build", true},
		{"build/out.js", false},
		{"src/build", true},
		{"root-only.txt", false},
		{"sub/root-only.txt", false},
		{"node_modules", true},
		{"sub/node_modules", true},
		{"docs/a/draft.md", false},
		{"docs/draft.md", false},
		{"x/generated", true},
		{"generated", true},
		{"temp1", false},
		{"temps", false},
		{"notes.tmp", false},
		{"src/notes.tmp", false},
		{"a/b/c.txt", false},
		{"z/a/b/c.txt", false},
		{"sub/local.txt", false},
		{"local.txt", false},
		{"src/main.go", false},
	}

	for _, p := range paths {
		full := filepath.Join(root, p.path)
		if p.isDir {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
		} else {
			write(t, full, "x")
		}
	}

	for _, cmd := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
	} {
		c := exec.Command("git", cmd...)
		c.Dir = root
		if out, err := c.CombinedOutput(); err != nil {
			t.Skipf("could not set up a git repo: %v: %s", err, out)
		}
	}

	st := stackFor(t, root, "sub")

	for _, p := range paths {
		c := exec.Command("git", "check-ignore", "-q", "--", p.path)
		c.Dir = root
		err := c.Run()

		wantIgnored := false
		if err == nil {
			wantIgnored = true
		} else if code := c.ProcessState.ExitCode(); code != 1 {
			t.Fatalf("git check-ignore failed for %q with exit %d", p.path, code)
		}

		got := st.Ignored(filepath.Join(root, p.path), p.isDir)
		if got != wantIgnored {
			t.Errorf("%q: got ignored=%v, git says %v", p.path, got, wantIgnored)
		}
	}
}
