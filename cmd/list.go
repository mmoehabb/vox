package cmd

import (
	"fmt"

	"vox/internal"

	"github.com/spf13/cobra"
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
		for _, entry := range words {
			fmt.Printf("%s: %s\n", entry.Word, entry.Definition)
		}
		return nil
	},
}