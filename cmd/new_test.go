package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runNew(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()

	newInto = "."
	newDryRun = false
	newForce = false
	for _, name := range []string{"into", "dry-run", "force"} {
		if flag := newCmd.Flags().Lookup(name); flag != nil {
			flag.Changed = false
		}
	}

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetIn(strings.NewReader(stdin))
	rootCmd.SetArgs(append([]string{"new"}, args...))
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetIn(nil)
	})

	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}

const sample = "app\n├── src\n│   └── main.go\n└── go.mod\n"

func TestNewFromStdin(t *testing.T) {
	dest := t.TempDir()

	stdout, _, err := runNew(t, sample, "-", "--into", dest)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if info, err := os.Stat(filepath.Join(dest, "app", "src")); err != nil || !info.IsDir() {
		t.Error("app/src should exist as a directory")
	}
	if _, err := os.Stat(filepath.Join(dest, "app", "src", "main.go")); err != nil {
		t.Error("app/src/main.go should exist")
	}
	if !strings.Contains(stdout, "2 directories, 2 files") {
		t.Errorf("expected a summary line:\n%s", stdout)
	}
}

func TestNewFromFile(t *testing.T) {
	dest := t.TempDir()
	input := filepath.Join(t.TempDir(), "tree.txt")
	if err := os.WriteFile(input, []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runNew(t, "", input, "--into", dest); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "app", "go.mod")); err != nil {
		t.Error("app/go.mod should exist")
	}
}

func TestNewDryRun(t *testing.T) {
	dest := t.TempDir()

	stdout, _, err := runNew(t, sample, "-", "--into", dest, "--dry-run")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "Would create") {
		t.Errorf("dry run should say what it would do:\n%s", stdout)
	}

	found, err := os.ReadDir(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 0 {
		t.Errorf("dry run wrote to disk: %v", found)
	}
}

func TestNewSkipsExistingThenForces(t *testing.T) {
	dest := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dest, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(dest, "app", "go.mod")
	if err := os.WriteFile(existing, []byte("module x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, err := runNew(t, sample, "-", "--into", dest)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stderr, "already exists") {
		t.Errorf("expected a skip notice:\n%s", stderr)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "module x" {
		t.Fatal("the existing file was modified without --force")
	}

	if _, _, err := runNew(t, sample, "-", "--into", dest, "--force"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	data, err = os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Error("--force should have truncated the file")
	}
}

func TestNewRejectsEmptyInput(t *testing.T) {
	if _, _, err := runNew(t, "\n\n", "-", "--into", t.TempDir()); err == nil {
		t.Error("expected an error for input with no tree in it")
	}
}

func TestNewMissingFile(t *testing.T) {
	if _, _, err := runNew(t, "", filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Error("expected an error for a missing input file")
	}
}

func TestInitWritesConfig(t *testing.T) {
	dest := t.TempDir()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init", dest})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, ".treegenrc")); err != nil {
		t.Error(".treegenrc should have been written")
	}

	rootCmd.SetArgs([]string{"init", dest})
	if err := rootCmd.Execute(); err == nil {
		t.Error("init should refuse to overwrite an existing config")
	}
}
