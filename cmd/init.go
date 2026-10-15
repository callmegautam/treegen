/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/callmegautam/treegen/internal/config"
)

var initCmd = &cobra.Command{
	Use:           "init [dir]",
	Short:         "Write a starter .treegenrc in the current directory",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}

		path := filepath.Join(dir, config.Names[0])
		if err := config.Write(path); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
