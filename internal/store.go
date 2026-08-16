package internal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrNotFound = errors.New("word not found")
	ErrCorrupt  = errors.New("corrupt data file")
)

const (
	defaultDir  = ".local/share/vox"
	defaultFile = "words.json"
)

func dataPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, defaultDir)
	} else {
		base = filepath.Join(base, "vox")
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(base, defaultFile), nil
}

func Load() ([]WordEntry, error) {
	path, err := dataPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []WordEntry{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []WordEntry{}, nil
	}
	var words []WordEntry
	if err := json.Unmarshal(data, &words); err != nil {
		return nil, ErrCorrupt
	}
	return words, nil
}

func Save(words []WordEntry) error {
	path, err := dataPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(words, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func AddOrReplace(entry WordEntry) error {
	words, err := Load()
	if err != nil {
		return err
	}
	for i, w := range words {
		if w.Word == entry.Word {
			words[i] = entry
			return Save(words)
		}
	}
	words = append(words, entry)
	return Save(words)
}

func Delete(word string) error {
	words, err := Load()
	if err != nil {
		return err
	}
	for i, w := range words {
		if w.Word == word {
			words = append(words[:i], words[i+1:]...)
			return Save(words)
		}
	}
	return ErrNotFound
}

func Get(word string) (WordEntry, error) {
	words, err := Load()
	if err != nil {
		return WordEntry{}, err
	}
	for _, w := range words {
		if w.Word == word {
			return w, nil
		}
	}
	return WordEntry{}, ErrNotFound
}

func FindPrefix(prefix string) []WordEntry {
	words, _ := Load()
	var matches []WordEntry
	for _, w := range words {
		if len(w.Word) >= len(prefix) && w.Word[:len(prefix)] == prefix {
			matches = append(matches, w)
		}
	}
	return matches
}

func FindFuzzy(pattern string) []WordEntry {
	words, _ := Load()
	var matches []WordEntry
	for _, w := range words {
		if containsFuzzy(w.Word, pattern) {
			matches = append(matches, w)
		}
	}
	return matches
}

func containsFuzzy(s, pattern string) bool {
	if pattern == "" {
		return true
	}
	i := 0
	for j := 0; j < len(s) && i < len(pattern); j++ {
		if s[j] == pattern[i] {
			i++
		}
	}
	return i == len(pattern)
}

func Import(entries []WordEntry, override bool) error {
	words, err := Load()
	if err != nil {
		return err
	}
	existing := make(map[string]int)
	for i, w := range words {
		existing[w.Word] = i
	}
	for _, e := range entries {
		if idx, ok := existing[e.Word]; ok {
			if override {
				words[idx] = e
			}
		} else {
			words = append(words, e)
			existing[e.Word] = len(words) - 1
		}
	}
	return Save(words)
}

func All() ([]WordEntry, error) {
	return Load()
}