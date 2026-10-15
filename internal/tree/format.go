/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package tree

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "text", "txt", "tree":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unknown format %q, expected text, json or yaml", s)
	}
}

type exported struct {
	Name     string      `json:"name" yaml:"name"`
	Type     string      `json:"type" yaml:"type"`
	Target   string      `json:"target,omitempty" yaml:"target,omitempty"`
	Error    string      `json:"error,omitempty" yaml:"error,omitempty"`
	Children []*exported `json:"children,omitempty" yaml:"children,omitempty"`
}

func export(node *Node) *exported {
	out := &exported{
		Name:   node.Name,
		Type:   "file",
		Target: node.Target,
	}

	switch {
	case node.IsSymlink:
		out.Type = "symlink"
	case node.IsDir:
		out.Type = "directory"
	}

	if node.Err != nil {
		out.Error = node.Err.Error()
	}

	for _, child := range node.Children {
		out.Children = append(out.Children, export(child))
	}

	return out
}

func marshal(root *Node, format Format) (string, error) {
	payload := export(root)

	switch format {
	case FormatJSON:
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	case FormatYAML:
		data, err := yaml.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("format %q is not a marshalled format", format)
	}
}
