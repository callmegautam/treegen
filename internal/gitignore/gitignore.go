/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package gitignore

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type pattern struct {
	re      *regexp.Regexp
	negate  bool
	dirOnly bool
}

type Set struct {
	patterns []pattern
}

type Stack struct {
	sets []*Set
	dirs []string
}

func Parse(r io.Reader) (*Set, error) {
	set := &Set{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		p, ok := compile(scanner.Text())
		if ok {
			set.patterns = append(set.patterns, p)
		}
	}

	return set, scanner.Err()
}

func ParseFile(path string) (*Set, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func (s *Set) Empty() bool {
	return s == nil || len(s.patterns) == 0
}

func (s *Set) match(rel string, isDir bool) (bool, bool) {
	if s.Empty() {
		return false, false
	}

	ignored := false
	matched := false

	for _, p := range s.patterns {
		if p.dirOnly && !isDir {
			continue
		}
		if !p.re.MatchString(rel) {
			continue
		}
		matched = true
		ignored = !p.negate
	}

	return ignored, matched
}

func NewStack() *Stack {
	return &Stack{}
}

func (st *Stack) Push(dir string, set *Set) {
	if set.Empty() {
		return
	}
	st.sets = append(st.sets, set)
	st.dirs = append(st.dirs, filepath.Clean(dir))
}

func (st *Stack) PushDir(dir string) error {
	set, err := ParseFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	st.Push(dir, set)
	return nil
}

func (st *Stack) Clone() *Stack {
	out := &Stack{
		sets: make([]*Set, len(st.sets)),
		dirs: make([]string, len(st.dirs)),
	}
	copy(out.sets, st.sets)
	copy(out.dirs, st.dirs)
	return out
}

func (st *Stack) Ignored(path string, isDir bool) bool {
	if st == nil || len(st.sets) == 0 {
		return false
	}

	path = filepath.Clean(path)
	base := st.dirs[0]

	rel, err := filepath.Rel(base, path)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return false
	}

	segments := strings.Split(filepath.ToSlash(rel), "/")

	for i := 1; i < len(segments); i++ {
		ancestor := filepath.Join(base, filepath.FromSlash(strings.Join(segments[:i], "/")))
		if st.matchOne(ancestor, true) {
			return true
		}
	}

	return st.matchOne(path, isDir)
}

func (st *Stack) matchOne(path string, isDir bool) bool {
	ignored := false

	for i := range st.sets {
		rel, err := filepath.Rel(st.dirs[i], path)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}

		if result, matched := st.sets[i].match(filepath.ToSlash(rel), isDir); matched {
			ignored = result
		}
	}

	return ignored
}

func compile(line string) (pattern, bool) {
	raw := strings.TrimRight(line, " \t")
	if raw == "" || strings.HasPrefix(raw, "#") {
		return pattern{}, false
	}

	p := pattern{}

	if strings.HasPrefix(raw, "!") {
		p.negate = true
		raw = raw[1:]
	}

	raw = strings.TrimPrefix(raw, `\`)

	if strings.HasSuffix(raw, "/") {
		p.dirOnly = true
		raw = strings.TrimSuffix(raw, "/")
	}

	if raw == "" {
		return pattern{}, false
	}

	anchored := strings.Contains(raw, "/")
	raw = strings.TrimPrefix(raw, "/")

	var b strings.Builder
	b.WriteString("^")
	if !anchored {
		b.WriteString("(?:.*/)?")
	}
	b.WriteString(translate(raw))
	b.WriteString("(?:/.*)?$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return pattern{}, false
	}

	p.re = re
	return p, true
}

func translate(glob string) string {
	var b strings.Builder

	for i := 0; i < len(glob); i++ {
		switch c := glob[i]; c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				i++
				switch {
				case i+1 < len(glob) && glob[i+1] == '/':
					i++
					b.WriteString("(?:.*/)?")
				case i == len(glob)-1:
					b.WriteString(".*")
				default:
					b.WriteString(".*")
				}
				continue
			}
			b.WriteString("[^/]*")
		case '?':
			b.WriteString("[^/]")
		case '[':
			end := i + 1
			if end < len(glob) && (glob[end] == '!' || glob[end] == '^') {
				end++
			}
			if end < len(glob) && glob[end] == ']' {
				end++
			}
			for end < len(glob) && glob[end] != ']' {
				end++
			}
			if end >= len(glob) {
				b.WriteString(regexp.QuoteMeta(string(c)))
				continue
			}
			class := glob[i+1 : end]
			if strings.HasPrefix(class, "!") {
				class = "^" + class[1:]
			}
			b.WriteString("[" + class + "]")
			i = end
		case '\\':
			if i+1 < len(glob) {
				i++
				b.WriteString(regexp.QuoteMeta(string(glob[i])))
				continue
			}
			b.WriteString(regexp.QuoteMeta(string(c)))
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}

	return b.String()
}
