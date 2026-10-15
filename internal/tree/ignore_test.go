package tree

import "testing"

func TestDefaultMatcher(t *testing.T) {
	m := DefaultMatcher()

	for _, name := range DefaultIgnores {
		if !m.Match(name) {
			t.Errorf("expected %q to be ignored by default", name)
		}
	}

	for _, name := range []string{"main.go", "cmd", "gitignore", "modules"} {
		if m.Match(name) {
			t.Errorf("did not expect %q to be ignored", name)
		}
	}
}

func TestMatcherAddExtendsDefaults(t *testing.T) {
	m := DefaultMatcher()
	m.Add("dist", " build ")

	if !m.Match("dist") || !m.Match("build") {
		t.Error("added patterns should match, with surrounding space trimmed")
	}
	if !m.Match(".git") {
		t.Error("adding patterns must not drop the defaults")
	}
}

func TestMatcherGlob(t *testing.T) {
	m := NewMatcher([]string{"*.log", "tmp?"})

	cases := map[string]bool{
		"debug.log": true,
		"log":       false,
		"tmp1":      true,
		"tmp":       false,
		"tmpdir":    false,
	}

	for name, want := range cases {
		if got := m.Match(name); got != want {
			t.Errorf("Match(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestMatcherIgnoresEmptyPatterns(t *testing.T) {
	m := NewMatcher([]string{"", "   ", "dist"})

	if len(m.Patterns()) != 1 {
		t.Fatalf("expected empty patterns to be dropped, got %v", m.Patterns())
	}
	if m.Match("") {
		t.Error("empty name should not match")
	}
}

func TestNilMatcher(t *testing.T) {
	var m *Matcher
	if m.Match("anything") {
		t.Error("nil matcher should match nothing")
	}
}

func TestDefaultMatcherIsIndependent(t *testing.T) {
	a := DefaultMatcher()
	a.Add("dist")

	if DefaultMatcher().Match("dist") {
		t.Error("mutating one matcher must not affect the default ignore list")
	}
}
