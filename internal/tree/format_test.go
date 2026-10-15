package tree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseFormat(t *testing.T) {
	cases := map[string]Format{
		"":     FormatText,
		"text": FormatText,
		"TEXT": FormatText,
		" txt": FormatText,
		"json": FormatJSON,
		"yaml": FormatYAML,
		"yml":  FormatYAML,
	}

	for in, want := range cases {
		got, err := ParseFormat(in)
		if err != nil {
			t.Errorf("ParseFormat(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseFormat(%q) = %q, want %q", in, got, want)
		}
	}

	if _, err := ParseFormat("toml"); err == nil {
		t.Error("expected an error for an unknown format")
	}
}

func formatFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"cmd/root.go", "main.go"} {
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

func TestJSONFormat(t *testing.T) {
	root := formatFixture(t)

	res, err := Generate(Config{Root: root, Format: FormatJSON})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var payload exported
	if err := json.Unmarshal([]byte(res.Text), &payload); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}

	if payload.Type != "directory" || payload.Name != filepath.Base(root) {
		t.Errorf("unexpected root: %+v", payload)
	}
	if len(payload.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(payload.Children))
	}
	if payload.Children[0].Name != "cmd" || payload.Children[0].Type != "directory" {
		t.Errorf("expected the directory first, got %+v", payload.Children[0])
	}
	if payload.Children[0].Children[0].Name != "root.go" {
		t.Error("nested children should be present")
	}
}

func TestYAMLFormat(t *testing.T) {
	root := formatFixture(t)

	res, err := Generate(Config{Root: root, Format: FormatYAML})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var payload exported
	if err := yaml.Unmarshal([]byte(res.Text), &payload); err != nil {
		t.Fatalf("output is not valid yaml: %v", err)
	}
	if payload.Name != filepath.Base(root) || len(payload.Children) != 2 {
		t.Errorf("unexpected payload: %+v", payload)
	}
}

func TestFormatsIgnoreTheFence(t *testing.T) {
	root := formatFixture(t)

	res, err := Generate(Config{Root: root, Format: FormatJSON, Fence: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(res.Text, "```") {
		t.Error("json output must not be fenced")
	}
}

func TestSymlinkInJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "real.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "real.txt"), filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	res, err := Generate(Config{Root: root, Format: FormatJSON})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var payload exported
	if err := json.Unmarshal([]byte(res.Text), &payload); err != nil {
		t.Fatal(err)
	}

	var link *exported
	for _, c := range payload.Children {
		if c.Name == "link" {
			link = c
		}
	}
	if link == nil || link.Type != "symlink" || link.Target == "" {
		t.Errorf("symlink should carry its type and target, got %+v", link)
	}
}
