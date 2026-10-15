/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/callmegautam/treegen/internal/gitignore"
)

type Node struct {
	Name      string
	Children  []*Node
	IsDir     bool
	IsSymlink bool
	Target    string
	Err       error
}

type Result struct {
	Text     string
	Warnings []error
}

type Config struct {
	Root      string
	Depth     int
	Ignore    *Matcher
	Fence     bool
	Format    Format
	GitIgnore bool
}

type walker struct {
	cfg Config
	sem chan struct{}
}

func Generate(cfg Config) (Result, error) {
	root := cfg.Root
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		root = wd
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return Result{}, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return Result{}, err
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("%s is not a directory", abs)
	}

	if cfg.Ignore == nil {
		cfg.Ignore = DefaultMatcher()
	}

	w := &walker{
		cfg: cfg,
		sem: make(chan struct{}, runtime.NumCPU()),
	}

	ignores := gitignore.NewStack()
	if cfg.GitIgnore {
		if err := ignores.PushDir(abs); err != nil {
			return Result{}, err
		}
	}

	rootNode := w.build(abs, 0, ignores)

	if cfg.Format != "" && cfg.Format != FormatText {
		text, err := marshal(rootNode, cfg.Format)
		if err != nil {
			return Result{}, err
		}
		return Result{Text: text, Warnings: collectErrors(rootNode, rootName(abs))}, nil
	}

	var builder strings.Builder
	if cfg.Fence {
		builder.WriteString("```text\n")
	}

	builder.WriteString(rootName(abs) + "\n")
	for i, child := range rootNode.Children {
		render(&builder, child, "", i == len(rootNode.Children)-1)
	}

	if cfg.Fence {
		builder.WriteString("```\n")
	}

	return Result{Text: builder.String(), Warnings: collectErrors(rootNode, rootName(abs))}, nil
}

func rootName(abs string) string {
	name := filepath.Base(abs)
	if name == "." || name == string(filepath.Separator) {
		return abs
	}
	return name
}

func (w *walker) build(path string, level int, ignores *gitignore.Stack) *Node {
	info, err := os.Lstat(path)
	if err != nil {
		return &Node{Name: filepath.Base(path), Err: err}
	}

	node := &Node{
		Name:      info.Name(),
		IsSymlink: info.Mode()&os.ModeSymlink != 0,
		IsDir:     info.IsDir(),
	}

	if node.IsSymlink {
		node.Target, _ = os.Readlink(path)
		if node.Target == "" {
			node.Target = "?"
		}
		return node
	}

	if !node.IsDir {
		return node
	}

	if w.cfg.Depth > 0 && level >= w.cfg.Depth {
		return node
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		node.Err = err
		return node
	}

	if w.cfg.GitIgnore && level > 0 {
		scoped := ignores.Clone()
		if err := scoped.PushDir(path); err == nil {
			ignores = scoped
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	appendChild := func(child *Node) {
		mu.Lock()
		node.Children = append(node.Children, child)
		mu.Unlock()
	}

	for _, entry := range entries {
		if w.cfg.Ignore.Match(entry.Name()) {
			continue
		}

		childPath := filepath.Join(path, entry.Name())

		if w.cfg.GitIgnore && ignores.Ignored(childPath, entry.IsDir()) {
			continue
		}

		select {
		case w.sem <- struct{}{}:
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-w.sem }()
				appendChild(w.build(childPath, level+1, ignores))
			}()
		default:
			appendChild(w.build(childPath, level+1, ignores))
		}
	}

	wg.Wait()
	sortNodes(node.Children)

	return node
}

func sortNodes(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		a := nodes[i]
		b := nodes[j]

		aDir := a.IsDir && !a.IsSymlink
		bDir := b.IsDir && !b.IsSymlink

		if aDir != bDir {
			return aDir
		}

		if strings.EqualFold(a.Name, b.Name) {
			return a.Name < b.Name
		}

		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
}

func render(builder *strings.Builder, node *Node, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	builder.WriteString(prefix + connector + node.Name)
	if node.IsSymlink {
		builder.WriteString(" -> " + node.Target)
	}
	if node.Err != nil {
		builder.WriteString(" [unreadable]")
	}
	builder.WriteString("\n")

	newPrefix := prefix
	if isLast {
		newPrefix += "    "
	} else {
		newPrefix += "│   "
	}

	for i, child := range node.Children {
		render(builder, child, newPrefix, i == len(node.Children)-1)
	}
}

func collectErrors(root *Node, rootLabel string) []error {
	var errs []error
	var walk func(n *Node, path string)
	walk = func(n *Node, path string) {
		if n.Err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, n.Err))
		}
		for _, c := range n.Children {
			walk(c, filepath.Join(path, c.Name))
		}
	}
	walk(root, rootLabel)
	return errs
}
