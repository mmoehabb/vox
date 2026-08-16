package cmd

import (
	"fmt"
	"strings"

	"github.com/mmoehabb/vox/internal"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <word>",
	Short: "Delete a word entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		word := strings.TrimSpace(args[0])
		if err := internal.Delete(word); err != nil {
			if err == internal.ErrNotFound {
				return fmt.Errorf("word not found: %s", word)
			}
			return err
		}
		fmt.Printf("Deleted: %s\n", word)
		return nil
	},
}