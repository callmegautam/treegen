/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	clipboardutil "github.com/callmegautam/treegen/internal/clipboard"
	"github.com/callmegautam/treegen/internal/config"
	"github.com/callmegautam/treegen/internal/tree"
)

var version = "dev"

var (
	depth        int
	ignore       string
	output       string
	format       string
	fence        bool
	noClipboard  bool
	quiet        bool
	useGitIgnore bool
	configPath   string
	noConfig     bool
)

var rootCmd = &cobra.Command{
	Use:           "treegen [path]",
	Short:         "Generate folder tree and copy to clipboard",
	Version:       version,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		root := "."
		if len(args) == 1 {
			root = args[0]
		}

		settings, err := resolve(cmd, root)
		if err != nil {
			return err
		}

		parsedFormat, err := tree.ParseFormat(settings.format)
		if err != nil {
			return err
		}

		matcher := tree.DefaultMatcher()
		if settings.ignore != "" {
			matcher.Add(strings.Split(settings.ignore, ",")...)
		}

		result, err := tree.Generate(tree.Config{
			Root:      root,
			Depth:     settings.depth,
			Ignore:    matcher,
			Fence:     settings.fence,
			Format:    parsedFormat,
			GitIgnore: settings.gitignore,
		})
		if err != nil {
			return err
		}

		stdout := cmd.OutOrStdout()
		stderr := cmd.ErrOrStderr()

		if output == "" || !quiet {
			fmt.Fprint(stdout, result.Text)
		}

		for _, warn := range result.Warnings {
			fmt.Fprintf(stderr, "warning: %v\n", warn)
		}

		if output != "" {
			if err := os.WriteFile(output, []byte(result.Text), 0o644); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(stdout, "Written to %s\n", output)
			}
		}

		if settings.clipboard && output == "" {
			if err := clipboardutil.CopyToClipboard(result.Text); err != nil {
				fmt.Fprintf(stderr, "warning: could not copy to clipboard: %v\n", err)
			} else if !quiet {
				fmt.Fprintln(stdout, "Copied to clipboard.")
			}
		}

		if !quiet {
			fmt.Fprintf(stdout, "Generated in %v\n", time.Since(start).Round(time.Millisecond))
		}

		return nil
	},
}

type settings struct {
	depth     int
	ignore    string
	format    string
	fence     bool
	gitignore bool
	clipboard bool
}

func resolve(cmd *cobra.Command, root string) (settings, error) {
	s := settings{
		depth:     depth,
		ignore:    ignore,
		format:    format,
		fence:     fence,
		gitignore: useGitIgnore,
		clipboard: !noClipboard,
	}

	if noConfig {
		return s, nil
	}

	var file *config.Config
	var err error

	if configPath != "" {
		file, err = config.LoadFile(configPath)
	} else {
		file, err = config.Load(root)
		if errors.Is(err, config.ErrNotFound) {
			return s, nil
		}
	}
	if err != nil {
		return s, err
	}

	flags := cmd.Flags()

	if file.Depth != nil && !flags.Changed("depth") {
		s.depth = *file.Depth
	}
	if len(file.Ignore) > 0 {
		merged := append([]string{}, file.Ignore...)
		if s.ignore != "" {
			merged = append(merged, s.ignore)
		}
		s.ignore = strings.Join(merged, ",")
	}
	if file.Format != nil && !flags.Changed("format") {
		s.format = *file.Format
	}
	if file.Fence != nil && !flags.Changed("fence") {
		s.fence = *file.Fence
	}
	if file.GitIgnore != nil && !flags.Changed("gitignore") {
		s.gitignore = *file.GitIgnore
	}
	if file.Clipboard != nil && !flags.Changed("no-clipboard") {
		s.clipboard = *file.Clipboard
	}

	return s, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "treegen: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVar(&depth, "depth", 0, "Limit folder depth (0 = unlimited)")
	rootCmd.Flags().StringVar(&ignore, "ignore", "", "Comma-separated names or globs to ignore, in addition to the defaults")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Write the tree to a file instead of the clipboard")
	rootCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json or yaml")
	rootCmd.Flags().BoolVar(&fence, "fence", false, "Wrap the output in a markdown code fence")
	rootCmd.Flags().BoolVar(&useGitIgnore, "gitignore", false, "Also skip anything the repository's .gitignore files skip")
	rootCmd.Flags().BoolVar(&noClipboard, "no-clipboard", false, "Print the tree without copying it")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Print only the tree, no status lines")
	rootCmd.Flags().StringVar(&configPath, "config", "", "Path to a config file, overriding the search")
	rootCmd.Flags().BoolVar(&noConfig, "no-config", false, "Ignore any config file")
}
