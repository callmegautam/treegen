/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package tree

import (
	"path/filepath"
	"strings"
)

var DefaultIgnores = []string{
	".git",
	"node_modules",
	".DS_Store",
	".idea",
	".vscode",
}

type Matcher struct {
	patterns []string
}

func NewMatcher(patterns []string) *Matcher {
	m := &Matcher{}
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			m.patterns = append(m.patterns, p)
		}
	}
	return m
}

func DefaultMatcher() *Matcher {
	return NewMatcher(DefaultIgnores)
}

func (m *Matcher) Add(patterns ...string) {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			m.patterns = append(m.patterns, p)
		}
	}
}

func (m *Matcher) Match(name string) bool {
	if m == nil {
		return false
	}
	for _, p := range m.patterns {
		if p == name {
			return true
		}
		if ok, err := filepath.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
}

func (m *Matcher) Patterns() []string {
	out := make([]string, len(m.patterns))
	copy(out, m.patterns)
	return out
}
