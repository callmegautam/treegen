/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Names = []string{".treegenrc", ".treegenrc.yaml", ".treegenrc.yml", "treegen.yaml", "treegen.yml"}

type Config struct {
	Depth     *int     `yaml:"depth,omitempty"`
	Ignore    []string `yaml:"ignore,omitempty"`
	GitIgnore *bool    `yaml:"gitignore,omitempty"`
	Format    *string  `yaml:"format,omitempty"`
	Fence     *bool    `yaml:"fence,omitempty"`
	Clipboard *bool    `yaml:"clipboard,omitempty"`

	Path string `yaml:"-"`
}

var ErrNotFound = errors.New("no config file found")

func Find(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		for _, name := range Names {
			candidate := filepath.Join(dir, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

func Load(start string) (*Config, error) {
	path, err := Find(start)
	if err != nil {
		return nil, err
	}
	return LoadFile(path)
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	cfg.Path = path
	return cfg, nil
}

const Template = `# treegen configuration
# Flags given on the command line always win over this file.

# Limit how deep the tree goes, 0 means unlimited.
depth: 0

# Names or globs to skip, on top of the built in defaults
# (.git, node_modules, .DS_Store, .idea, .vscode).
ignore:
  - dist
  - build
  - "*.log"

# Skip anything the repository's .gitignore files already skip.
gitignore: true

# text, json or yaml.
format: text

# Wrap text output in a markdown code fence.
fence: false

# Copy the result to the clipboard.
clipboard: true
`

func Write(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	return os.WriteFile(path, []byte(Template), 0o644)
}
