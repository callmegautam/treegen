/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var depth int
var ignore string

var rootCmd = &cobra.Command{
	Use:   "treegen",
	Short: "Generate folder tree and copy to clipboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
func init() {
	rootCmd.Flags().IntVar(&depth, "depth", 0, "Limit folder depth (0 = unlimited)")
	rootCmd.Flags().StringVar(&ignore, "ignore", "", "Comma-separated folders to ignore")
}
