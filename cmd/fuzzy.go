package cmd

import (
	"fmt"
	"strings"

	"vox/internal"

	"github.com/spf13/cobra"
)

var fuzzyCmd = &cobra.Command{
	Use:   "fuzzy <pattern>",
	Short: "Fuzzy search words by pattern",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := strings.TrimSpace(args[0])
		matches := internal.FindFuzzy(pattern)
		if len(matches) == 0 {
			fmt.Println("No matches found")
			return nil
		}
		for i, entry := range matches {
			if i > 0 {
				fmt.Println()
			}
			printEntry(entry)
		}
		return nil
	},
}