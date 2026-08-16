package cmd

import (
	"fmt"
	"strings"

	"github.com/mmoehabb/vox/internal"

	"github.com/spf13/cobra"
)

var (
	addDefinition   string
	addReferenceURL string
	addExamples     []string
	addLinks        []string
)

var addCmd = &cobra.Command{
	Use:   "add <word>",
	Short: "Add a new word entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		word := strings.TrimSpace(args[0])
		if word == "" {
			return internal.ErrEmptyWord
		}

		definition := strings.TrimSpace(addDefinition)
		if definition == "" {
			return internal.ErrEmptyDefinition
		}

		var refURL *string
		if strings.TrimSpace(addReferenceURL) != "" {
			u := strings.TrimSpace(addReferenceURL)
			refURL = &u
		}

		examples := make([]string, 0, len(addExamples))
		for _, e := range addExamples {
			if t := strings.TrimSpace(e); t != "" {
				examples = append(examples, t)
			}
		}

		links := make([]internal.Link, 0, len(addLinks))
		for _, l := range addLinks {
			parts := strings.SplitN(l, "|", 3)
			if len(parts) < 2 {
				return fmt.Errorf("link format: title|url|description")
			}
			title := strings.TrimSpace(parts[0])
			urlStr := strings.TrimSpace(parts[1])
			var desc *string
			if len(parts) == 3 && strings.TrimSpace(parts[2]) != "" {
				d := strings.TrimSpace(parts[2])
				desc = &d
			}
			links = append(links, internal.Link{
				Title:       title,
				URL:         urlStr,
				Description: desc,
			})
		}

		entry := internal.WordEntry{
			Word:         word,
			Definition:   definition,
			ReferenceURL: refURL,
			Examples:     examples,
			Links:        links,
		}

		if err := internal.ValidateWordEntry(entry); err != nil {
			return err
		}

		if err := internal.AddOrReplace(entry); err != nil {
			return err
		}

		fmt.Printf("Added/updated: %s\n", word)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDefinition, "definition", "d", "", "Definition (required)")
	addCmd.Flags().StringVarP(&addReferenceURL, "reference", "r", "", "Reference URL")
	addCmd.Flags().StringSliceVarP(&addExamples, "example", "e", nil, "Example sentences (repeatable)")
	addCmd.Flags().StringSliceVarP(&addLinks, "link", "l", nil, "Links as title|url|description (repeatable)")
	addCmd.MarkFlagRequired("definition")
}