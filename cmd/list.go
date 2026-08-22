package cmd

import (
	"fmt"

	"github.com/mmoehabb/vox/internal"

	"github.com/spf13/cobra"
)

var (
	listOffset int
	listLimit  int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all words",
	RunE: func(cmd *cobra.Command, args []string) error {
		words, err := internal.All()
		if err != nil {
			return err
		}
		if len(words) == 0 {
			fmt.Println("No words stored")
			return nil
		}

		start := listOffset
		if start > len(words) {
			start = len(words)
		}

		end := len(words)
		if listLimit > 0 && start+listLimit < end {
			end = start + listLimit
		}

		words = words[start:end]

		for _, entry := range words {
			def := entry.Definition
			if len(def) > 64 {
				def = def[:61] + "..."
			}
			fmt.Printf("*%s*: %s\n", entry.Word, def)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Number of entries to skip")
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 50, "Maximum entries to display (default: 50, 0 = all)")
}