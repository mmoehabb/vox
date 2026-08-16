package internal

import (
	"time"
)

type Link struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	URL         string  `json:"url"`
}

type WordEntry struct {
	Word          string   `json:"word"`
	Definition    string   `json:"definition"`
	ReferenceURL  *string  `json:"reference_url,omitempty"`
	Examples      []string `json:"examples,omitempty"`
	Links         []Link   `json:"links,omitempty"`
}

type ExportWrapper struct {
	Version     string      `json:"version"`
	ExportedAt  time.Time   `json:"exported_at"`
	Words       []WordEntry `json:"words"`
}