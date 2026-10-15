package config

import (
	"errors"
	"os"
	"path/filepath"
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

func TestLoadReadsEveryField(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".treegenrc"), `
depth: 3
ignore:
  - dist
  - "*.log"
gitignore: true
format: json
fence: true
clipboard: false
`)

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Depth == nil || *cfg.Depth != 3 {
		t.Errorf("depth = %v", cfg.Depth)
	}
	if len(cfg.Ignore) != 2 || cfg.Ignore[0] != "dist" {
		t.Errorf("ignore = %v", cfg.Ignore)
	}
	if cfg.GitIgnore == nil || !*cfg.GitIgnore {
		t.Error("gitignore should be true")
	}
	if cfg.Format == nil || *cfg.Format != "json" {
		t.Errorf("format = %v", cfg.Format)
	}
	if cfg.Fence == nil || !*cfg.Fence {
		t.Error("fence should be true")
	}
	if cfg.Clipboard == nil || *cfg.Clipboard {
		t.Error("clipboard should be false")
	}
}

func TestUnsetFieldsStayNil(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".treegenrc"), "depth: 2\n")

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Fence != nil || cfg.Format != nil || cfg.Clipboard != nil {
		t.Error("fields absent from the file must stay nil so flags and defaults win")
	}
}

func TestSearchesUpward(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".treegenrc"), "depth: 1\n")
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(deep)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Path != filepath.Join(root, ".treegenrc") {
		t.Errorf("found %q", cfg.Path)
	}
}

func TestNearestFileWins(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".treegenrc"), "depth: 1\n")
	sub := filepath.Join(root, "sub")
	write(t, filepath.Join(sub, ".treegenrc"), "depth: 9\n")

	cfg, err := Load(sub)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Depth == nil || *cfg.Depth != 9 {
		t.Error("the closest config file should win")
	}
}

func TestAlternateNames(t *testing.T) {
	for _, name := range Names {
		root := t.TempDir()
		write(t, filepath.Join(root, name), "depth: 4\n")

		cfg, err := Load(root)
		if err != nil {
			t.Fatalf("Load with %s: %v", name, err)
		}
		if cfg.Depth == nil || *cfg.Depth != 4 {
			t.Errorf("%s was not read", name)
		}
	}
}

func TestNotFound(t *testing.T) {
	deep := filepath.Join(t.TempDir(), "a", "b")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(deep); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInvalidYaml(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".treegenrc"), "depth: [this is not an int\n")

	if _, err := Load(root); err == nil {
		t.Error("expected a parse error")
	}
}

func TestWriteRefusesToClobber(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".treegenrc")

	if err := Write(path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := Write(path); err == nil {
		t.Error("a second Write should refuse to overwrite")
	}
}

func TestTemplateIsValidAndComplete(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".treegenrc")
	if err := Write(path); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("the generated template must parse: %v", err)
	}
	if cfg.Depth == nil || cfg.Format == nil || cfg.GitIgnore == nil || cfg.Fence == nil || cfg.Clipboard == nil {
		t.Error("the template should set every documented field")
	}
}
