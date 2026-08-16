package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	oldEnv := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldEnv)

	entry := WordEntry{
		Word:       "test",
		Definition: "a test definition",
		Examples:   []string{"example 1", "example 2"},
		Links: []Link{
			{Title: "Link 1", URL: "https://example.com"},
		},
	}

	if err := AddOrReplace(entry); err != nil {
		t.Fatalf("AddOrReplace failed: %v", err)
	}

	words, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Word != "test" {
		t.Fatalf("word mismatch: %s", words[0].Word)
	}

	// Test Get
	got, err := Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Definition != "a test definition" {
		t.Fatalf("definition mismatch: %s", got.Definition)
	}

	// Test Delete
	if err := Delete("test"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	words, _ = Load()
	if len(words) != 0 {
		t.Fatalf("expected 0 words after delete, got %d", len(words))
	}

	// Test Delete not found
	if err := Delete("nonexistent"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestImportExport(t *testing.T) {
	tmpDir := t.TempDir()
	oldEnv := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldEnv)

	entries := []WordEntry{
		{Word: "apple", Definition: "a fruit"},
		{Word: "banana", Definition: "another fruit"},
	}

	if err := Import(entries, false); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Export
	words, err := All()
	if err != nil {
		t.Fatalf("All failed: %v", err)
	}

	wrapper := ExportWrapper{
		Version:   "1.0.0",
		ExportedAt: timeNow(),
		Words:     words,
	}

	exportPath := filepath.Join(tmpDir, "export.json")
	data, _ := json.MarshalIndent(wrapper, "", "  ")
	if err := os.WriteFile(exportPath, data, 0o644); err != nil {
		t.Fatalf("Write export failed: %v", err)
	}

	// Clear and re-import
	if err := Save([]WordEntry{}); err != nil {
		t.Fatalf("Save empty failed: %v", err)
	}

	// Re-import with --nooverride (default)
	importData, _ := os.ReadFile(exportPath)
	var importedWrapper ExportWrapper
	json.Unmarshal(importData, &importedWrapper)
	if err := Import(importedWrapper.Words, false); err != nil {
		t.Fatalf("Re-import failed: %v", err)
	}

	words, _ = All()
	if len(words) != 2 {
		t.Fatalf("expected 2 words after import, got %d", len(words))
	}

	// Test override
	overrideEntries := []WordEntry{
		{Word: "apple", Definition: "updated definition"},
	}
	if err := Import(overrideEntries, true); err != nil {
		t.Fatalf("Override import failed: %v", err)
	}
	words, _ = All()
	for _, w := range words {
		if w.Word == "apple" && w.Definition != "updated definition" {
			t.Fatalf("override didn't work: %s", w.Definition)
		}
	}
}

func TestFuzzySearch(t *testing.T) {
	tmpDir := t.TempDir()
	oldEnv := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldEnv)

	entries := []WordEntry{
		{Word: "apple", Definition: "fruit"},
		{Word: "application", Definition: "software"},
		{Word: "banana", Definition: "fruit"},
	}
	for _, e := range entries {
		AddOrReplace(e)
	}

	matches := FindFuzzy("app")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for 'app', got %d", len(matches))
	}

	matches = FindFuzzy("xyz")
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for 'xyz', got %d", len(matches))
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		entry   WordEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: WordEntry{
				Word:       "test",
				Definition: "a test",
			},
			wantErr: false,
		},
		{
			name: "empty word",
			entry: WordEntry{
				Word:       "",
				Definition: "a test",
			},
			wantErr: true,
		},
		{
			name: "empty definition",
			entry: WordEntry{
				Word:       "test",
				Definition: "",
			},
			wantErr: true,
		},
		{
			name: "invalid reference URL",
			entry: WordEntry{
				Word:          "test",
				Definition:    "a test",
				ReferenceURL:  strPtr("not-a-url"),
			},
			wantErr: true,
		},
		{
			name: "valid reference URL",
			entry: WordEntry{
				Word:          "test",
				Definition:    "a test",
				ReferenceURL:  strPtr("https://example.com"),
			},
			wantErr: false,
		},
		{
			name: "invalid link URL",
			entry: WordEntry{
				Word:       "test",
				Definition: "a test",
				Links: []Link{
					{Title: "Link", URL: "not-a-url"},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWordEntry(tc.entry)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateImportWrapper(t *testing.T) {
	wrapper := &ExportWrapper{
		Version: "1.0.0",
		Words: []WordEntry{
			{Word: "test", Definition: "a test"},
		},
	}
	if err := ValidateImportWrapper(wrapper); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	wrapper.Version = ""
	if err := ValidateImportWrapper(wrapper); err == nil {
		t.Error("expected error for missing version")
	}

	wrapper.Version = "1.0.0"
	wrapper.Words = []WordEntry{
		{Word: "", Definition: "a test"},
	}
	if err := ValidateImportWrapper(wrapper); err == nil {
		t.Error("expected error for invalid word entry")
	}
}

func strPtr(s string) *string {
	return &s
}

// Helper to avoid importing time in test file
func timeNow() time.Time {
	return time.Now().UTC()
}