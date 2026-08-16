package cmd

import (
	"fmt"
	"strings"

	"vox/internal"

	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query <word>",
	Short: "Query a word by exact match",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		word := strings.TrimSpace(args[0])
		entry, err := internal.Get(word)
		if err != nil {
			if err == internal.ErrNotFound {
				return fmt.Errorf("word not found: %s", word)
			}
			return err
		}
		printEntry(entry)
		return nil
	},
}