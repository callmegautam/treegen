/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package scaffold

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Entry struct {
	Path  string
	IsDir bool
}

type Plan struct {
	Created []Entry
	Skipped []Entry
}

type Options struct {
	DryRun bool
	Force  bool
}

var connectors = []string{"├── ", "└── ", "├──", "└──", "|-- ", "`-- ", "+-- ", "|--", "`--", "+--"}

var ErrEmpty = errors.New("no entries found in the input")

func Parse(r io.Reader) ([]Entry, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	type parsed struct {
		depth int
		name  string
		dir   bool
	}

	var lines []parsed

	for scanner.Scan() {
		raw := scanner.Text()

		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "```") {
			continue
		}

		depth, name, ok := split(raw)
		if !ok {
			continue
		}

		name, dir := clean(name)
		if name == "" {
			continue
		}

		lines = append(lines, parsed{depth: depth, name: name, dir: dir})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	first := -1
	for i, line := range lines {
		if line.depth > 0 {
			first = i
			break
		}
	}
	if first == -1 {
		return nil, ErrEmpty
	}

	start := first
	if first > 0 {
		start = first - 1
	}

	var entries []Entry
	stack := []string{}
	base := 0

	for i := start; i < len(lines); i++ {
		line := lines[i]

		if line.depth == 0 {
			if i > start {
				break
			}
			if line.name != "." {
				stack = append(stack, line.name)
				base = 1
				entries = append(entries, Entry{Path: line.name, IsDir: true})
			}
			continue
		}

		for len(stack) > base+line.depth-1 {
			stack = stack[:len(stack)-1]
		}

		stack = append(stack, line.name)

		isDir := line.dir
		if !isDir && i+1 < len(lines) && lines[i+1].depth > line.depth {
			isDir = true
		}

		entries = append(entries, Entry{Path: filepath.Join(stack...), IsDir: isDir})
	}

	if len(entries) == 0 {
		return nil, ErrEmpty
	}

	return entries, nil
}

func split(line string) (int, string, bool) {
	runes := []rune(line)

	best := -1
	width := 0

	for _, c := range connectors {
		idx := strings.Index(line, c)
		if idx < 0 {
			continue
		}
		at := len([]rune(line[:idx]))
		if best == -1 || at < best {
			best = at
			width = len([]rune(c))
		}
	}

	if best == -1 {
		if strings.TrimSpace(line) == "" {
			return 0, "", false
		}
		if len(runes) > 0 && (runes[0] == ' ' || runes[0] == '\t') {
			return 0, "", false
		}
		return 0, strings.TrimSpace(line), true
	}

	return best/4 + 1, strings.TrimSpace(string(runes[best+width:])), true
}

func clean(name string) (string, bool) {
	if idx := strings.Index(name, " -> "); idx >= 0 {
		name = name[:idx]
	}
	for _, marker := range []string{" #", "\t#", " //", "\t//"} {
		if idx := strings.Index(name, marker); idx >= 0 {
			name = name[:idx]
		}
	}

	name = strings.TrimSpace(name)
	name = strings.Trim(name, "`\"'")
	name = strings.TrimSpace(name)

	isDir := strings.HasSuffix(name, "/")
	name = strings.TrimSuffix(name, "/")

	if name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
		if name != "." && name != ".." {
			name = filepath.Base(filepath.FromSlash(name))
		}
	}

	return name, isDir
}

func Apply(root string, entries []Entry, opts Options) (*Plan, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	plan := &Plan{}

	for _, entry := range entries {
		target := filepath.Join(abs, entry.Path)

		rel, err := filepath.Rel(abs, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("refusing to write outside %s: %s", abs, entry.Path)
		}

		info, statErr := os.Lstat(target)
		exists := statErr == nil

		if exists && !entry.IsDir && !opts.Force {
			plan.Skipped = append(plan.Skipped, entry)
			continue
		}
		if exists && entry.IsDir && info.IsDir() {
			continue
		}

		plan.Created = append(plan.Created, entry)

		if opts.DryRun {
			continue
		}

		if entry.IsDir {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, nil, 0o644); err != nil {
			return nil, err
		}
	}

	return plan, nil
}
