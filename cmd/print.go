package cmd

import (
	"fmt"

	"github.com/mmoehabb/vox/internal"
)

func printEntry(entry internal.WordEntry) {
	fmt.Printf("Word: %s\n", entry.Word)
	if entry.Pronunciation != nil && *entry.Pronunciation != "" {
		fmt.Printf("Pronunciation: %s\n", *entry.Pronunciation)
	}
	fmt.Printf("Definition: %s\n", entry.Definition)
	if entry.ReferenceURL != nil && *entry.ReferenceURL != "" {
		fmt.Printf("Reference: %s\n", *entry.ReferenceURL)
	}
	if len(entry.Examples) > 0 {
		fmt.Println("Examples:")
		for _, ex := range entry.Examples {
			fmt.Printf("  - %s\n", ex)
		}
	}
	if len(entry.Links) > 0 {
		fmt.Println("Links:")
		for _, link := range entry.Links {
			fmt.Printf("  - %s: %s", link.Title, link.URL)
			if link.Description != nil && *link.Description != "" {
				fmt.Printf(" (%s)", *link.Description)
			}
			fmt.Println()
		}
	}
}
