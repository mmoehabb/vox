package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mmoehabb/vox/internal"

	"github.com/spf13/cobra"
)

var (
	importOverride bool
)

var importCmd = &cobra.Command{
	Use:   "import <input.json>",
	Short: "Import words from JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return err
		}

		var wrapper internal.ExportWrapper
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return fmt.Errorf("invalid JSON format: %w", err)
		}

		if err := internal.ValidateImportWrapper(&wrapper); err != nil {
			return err
		}

		if err := internal.Import(wrapper.Words, importOverride); err != nil {
			return err
		}

		action := "imported"
		if importOverride {
			action = "imported/updated"
		}
		fmt.Printf("%s %d words from %s\n", action, len(wrapper.Words), inputPath)
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importOverride, "override", false, "Override existing entries")
}