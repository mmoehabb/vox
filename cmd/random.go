package cmd

import (
	"fmt"
	"strconv"

	"github.com/mmoehabb/vox/internal"

	"github.com/spf13/cobra"
)

var randomCmd = &cobra.Command{
	Use:   "random [count]",
	Short: "Print random word(s)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		count := 1
		if len(args) == 1 {
			var err error
			count, err = strconv.Atoi(args[0])
			if err != nil || count < 1 {
				return fmt.Errorf("count must be a positive integer")
			}
		}
		words, err := internal.Random(count)
		if err != nil {
			return err
		}
		for _, w := range words {
			fmt.Println(w.Word)
		}
		return nil
	},
}