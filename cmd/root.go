/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/callmegautam/treegen/internal/tree"

	clipboardutil "github.com/callmegautam/treegen/internal/clipboard"
)

var depth int
var ignore string

var rootCmd = &cobra.Command{
	Use:   "treegen",
	Short: "Generate folder tree and copy to clipboard",

	RunE: func(cmd *cobra.Command, args []string) error {

		start := time.Now()

		root, _ := os.Getwd()

		ignoreMap := tree.DefaultIgnores
		if ignore != "" {
			custom := strings.SplitSeq(ignore, ",")
			for val := range custom {
				ignoreMap[strings.TrimSpace(val)] = true
			}
		}

		output, err := tree.Generate(tree.Config{
			Root:   root,
			Depth:  depth,
			Ignore: ignoreMap,
		})
		if err != nil {
			return err
		}

		fmt.Println(output)

		err = clipboardutil.CopyToClipboard(output)
		if err != nil {
			return err
		}

		fmt.Printf("Copied to clipboard.\n")
		fmt.Printf("Generated in %v\n", time.Since(start).Round(time.Millisecond))

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
