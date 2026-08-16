package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mmoehabb/vox/internal"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <output.json>",
	Short: "Export all words to JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := args[0]
		words, err := internal.All()
		if err != nil {
			return err
		}

		wrapper := internal.ExportWrapper{
			Version:    "1.0.0",
			ExportedAt: time.Now().UTC(),
			Words:      words,
		}

		data, err := json.MarshalIndent(wrapper, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(outputPath, data, 0o644); err != nil {
			return err
		}

		fmt.Printf("Exported %d words to %s\n", len(words), outputPath)
		return nil
	},
}