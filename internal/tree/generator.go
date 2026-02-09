/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package tree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Node struct {
	Name     string
	Children []*Node
	IsDir    bool
}

type Config struct {
	Root   string
	Depth  int
	Ignore map[string]bool
}

func Generate(cfg Config) (string, error) {
	rootNode, err := buildTree(cfg.Root, 0, cfg)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	// builder.WriteString("```text\n")
	// info, err := os.Stat(cfg.Root)
	// builder.WriteString(info.Name() + "\n")
	builder.WriteString("\n")

	for i, child := range rootNode.Children {
		render(&builder, child, "", i == len(rootNode.Children)-1)
	}

	// builder.WriteString("```\n")

	return builder.String(), nil
}

func buildTree(path string, level int, cfg Config) (*Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Name:  info.Name(),
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return node, nil
	}

	if cfg.Depth > 0 && level >= cfg.Depth {
		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, entry := range entries {
		if cfg.Ignore[entry.Name()] {
			continue
		}

		wg.Add(1)

		go func(e os.DirEntry) {
			defer wg.Done()

			childPath := filepath.Join(path, e.Name())
			childNode, err := buildTree(childPath, level+1, cfg)
			if err != nil {
				return
			}

			mu.Lock()
			node.Children = append(node.Children, childNode)
			mu.Unlock()

		}(entry)
	}

	wg.Wait()

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Name < node.Children[j].Name
	})

	return node, nil
}

func render(builder *strings.Builder, node *Node, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	builder.WriteString(prefix + connector + node.Name + "\n")

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
