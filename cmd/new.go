/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/callmegautam/treegen/internal/scaffold"
)

var (
	newInto   string
	newDryRun bool
	newForce  bool
)

var newCmd = &cobra.Command{
	Use:   "new [file]",
	Short: "Create folders and files from a tree, the reverse of generating one",
	Long: `Read a tree, the kind treegen prints, and create it on disk.

Pass a file, or pipe the tree on stdin. Entries ending in "/" are always
directories, entries with children are directories, everything else becomes an
empty file. Existing files are left alone unless --force is given.`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		var reader io.Reader

		if len(args) == 1 && args[0] != "-" {
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer file.Close()
			reader = file
		} else {
			in := cmd.InOrStdin()
			if file, ok := in.(*os.File); ok {
				stat, err := file.Stat()
				if err != nil {
					return err
				}
				if stat.Mode()&os.ModeCharDevice != 0 {
					return fmt.Errorf("no input: pass a file or pipe a tree on stdin")
				}
			}
			reader = in
		}

		entries, err := scaffold.Parse(reader)
		if err != nil {
			return err
		}

		plan, err := scaffold.Apply(newInto, entries, scaffold.Options{
			DryRun: newDryRun,
			Force:  newForce,
		})
		if err != nil {
			return err
		}

		stdout := cmd.OutOrStdout()
		verb := "Created"
		if newDryRun {
			verb = "Would create"
		}

		for _, entry := range plan.Created {
			suffix := ""
			if entry.IsDir {
				suffix = "/"
			}
			fmt.Fprintf(stdout, "%s %s%s\n", verb, entry.Path, suffix)
		}

		for _, entry := range plan.Skipped {
			fmt.Fprintf(cmd.ErrOrStderr(), "skipped %s, it already exists\n", entry.Path)
		}

		dirs := 0
		for _, entry := range plan.Created {
			if entry.IsDir {
				dirs++
			}
		}

		fmt.Fprintf(stdout, "\n%d directories, %d files, %d skipped\n",
			dirs, len(plan.Created)-dirs, len(plan.Skipped))

		return nil
	},
}

func init() {
	newCmd.Flags().StringVarP(&newInto, "into", "C", ".", "Directory to create the tree in")
	newCmd.Flags().BoolVar(&newDryRun, "dry-run", false, "Show what would be created without touching the disk")
	newCmd.Flags().BoolVar(&newForce, "force", false, "Overwrite files that already exist")
	rootCmd.AddCommand(newCmd)
}
