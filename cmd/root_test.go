package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	depth = 0
	ignore = ""
	output = ""
	format = "text"
	fence = false
	noClipboard = false
	quiet = false
	useGitIgnore = false
	configPath = ""
	noConfig = false

	rootCmd.SetArgs(nil)
	for _, name := range []string{"depth", "ignore", "output", "format", "fence", "gitignore", "no-clipboard", "quiet", "config", "no-config"} {
		if flag := rootCmd.Flags().Lookup(name); flag != nil {
			flag.Changed = false
		}
	}

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"cmd/root.go", "main.go", "node_modules/pkg/index.js", "dist/bundle.js"} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestRunsAgainstPathArgument(t *testing.T) {
	root := fixture(t)

	stdout, _, err := run(t, root, "--no-clipboard")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "main.go") {
		t.Errorf("expected the tree for the given path:\n%s", stdout)
	}
	if strings.Contains(stdout, "node_modules") {
		t.Error("default ignores should apply")
	}
}

func TestIgnoreFlagExtendsDefaults(t *testing.T) {
	root := fixture(t)

	stdout, _, err := run(t, root, "--no-clipboard", "--ignore", "dist")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(stdout, "dist") {
		t.Error("--ignore should drop the named directory")
	}
	if strings.Contains(stdout, "node_modules") {
		t.Error("--ignore should extend the defaults, not replace them")
	}
}

func TestQuietPrintsOnlyTheTree(t *testing.T) {
	root := fixture(t)

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, noise := range []string{"Copied to clipboard", "Generated in"} {
		if strings.Contains(stdout, noise) {
			t.Errorf("--quiet should suppress %q:\n%s", noise, stdout)
		}
	}
	if !strings.Contains(stdout, "main.go") {
		t.Error("--quiet should still print the tree")
	}
}

func TestOutputWritesFileAndSkipsClipboard(t *testing.T) {
	root := fixture(t)
	dest := filepath.Join(t.TempDir(), "tree.md")

	stdout, _, err := run(t, root, "--output", dest)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("expected the output file: %v", err)
	}
	if !strings.Contains(string(data), "main.go") {
		t.Errorf("output file should hold the tree, got:\n%s", data)
	}
	if strings.Contains(stdout, "Copied to clipboard") {
		t.Error("--output should not also copy to the clipboard")
	}
}

func TestOutputToUnwritablePathFails(t *testing.T) {
	root := fixture(t)

	if _, _, err := run(t, root, "--output", filepath.Join(root, "missing", "dir", "tree.md")); err == nil {
		t.Error("expected an error when the output file cannot be written")
	}
}

func TestMissingPathFails(t *testing.T) {
	if _, _, err := run(t, filepath.Join(t.TempDir(), "nope"), "--no-clipboard"); err == nil {
		t.Error("expected an error for a path that does not exist")
	}
}

func TestTooManyArgsFails(t *testing.T) {
	if _, _, err := run(t, ".", "extra", "--no-clipboard"); err == nil {
		t.Error("expected an error for more than one path argument")
	}
}

func TestFenceFlag(t *testing.T) {
	root := fixture(t)

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet", "--fence")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.HasPrefix(stdout, "```text\n") {
		t.Errorf("--fence should wrap the output:\n%s", stdout)
	}
}

func TestClipboardFailureIsAWarningNotAnError(t *testing.T) {
	root := fixture(t)

	t.Setenv("PATH", t.TempDir())
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	stdout, stderr, err := run(t, root)
	if err != nil {
		t.Fatalf("a missing clipboard must not fail the command, got: %v", err)
	}
	if !strings.Contains(stdout, "main.go") {
		t.Error("the tree should still be printed")
	}
	if strings.Contains(stdout, "Copied to clipboard") && stderr != "" {
		t.Error("either the copy worked or it warned, not both")
	}
}

func TestFormatFlag(t *testing.T) {
	root := fixture(t)

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet", "--format", "json")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("expected json:\n%s", stdout)
	}

	if _, _, err := run(t, root, "--no-clipboard", "--format", "toml"); err == nil {
		t.Error("expected an error for an unknown format")
	}
}

func TestGitIgnoreFlag(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("dist/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet", "--gitignore")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(stdout, "dist") {
		t.Errorf("--gitignore should drop dist:\n%s", stdout)
	}
}

func TestConfigFileIsRead(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, ".treegenrc"), []byte("depth: 1\nformat: json\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("the config file should have selected json:\n%s", stdout)
	}
}

func TestFlagsBeatTheConfigFile(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, ".treegenrc"), []byte("format: json\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet", "--format", "text")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("the flag should have won:\n%s", stdout)
	}
}

func TestNoConfigSkipsTheFile(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, ".treegenrc"), []byte("format: json\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet", "--no-config")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("--no-config should have ignored the file:\n%s", stdout)
	}
}

func TestConfigIgnoreMergesWithTheFlag(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, ".treegenrc"), []byte("ignore:\n  - dist\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := run(t, root, "--no-clipboard", "--quiet", "--ignore", "cmd")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(stdout, "dist") {
		t.Error("the config ignore list should still apply")
	}
	if strings.Contains(stdout, "cmd") {
		t.Error("the flag ignore list should also apply")
	}
}

func TestBrokenConfigIsAnError(t *testing.T) {
	root := fixture(t)
	if err := os.WriteFile(filepath.Join(root, ".treegenrc"), []byte("depth: [broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, root, "--no-clipboard"); err == nil {
		t.Error("a malformed config file should fail loudly, not silently")
	}
}
