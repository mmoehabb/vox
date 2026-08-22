package internal

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEmptyWord       = errors.New("word cannot be empty")
	ErrEmptyDefinition = errors.New("definition cannot be empty")
	ErrInvalidURL      = errors.New("invalid URL format")
)

func ValidateWordEntry(entry WordEntry) error {
	if strings.TrimSpace(entry.Word) == "" {
		return ErrEmptyWord
	}
	if strings.TrimSpace(entry.Definition) == "" {
		return ErrEmptyDefinition
	}
	if entry.Pronunciation != nil && strings.TrimSpace(*entry.Pronunciation) == "" {
		return errors.New("pronunciation cannot be empty if provided")
	}
	if entry.ReferenceURL != nil && strings.TrimSpace(*entry.ReferenceURL) != "" {
		if _, err := url.ParseRequestURI(*entry.ReferenceURL); err != nil {
			return ErrInvalidURL
		}
	}
	for _, link := range entry.Links {
		if strings.TrimSpace(link.Title) == "" {
			return errors.New("link title cannot be empty")
		}
		if strings.TrimSpace(link.URL) == "" {
			return errors.New("link URL cannot be empty")
		}
		if _, err := url.ParseRequestURI(link.URL); err != nil {
			return ErrInvalidURL
		}
	}
	return nil
}

func ValidateImportWrapper(wrapper *ExportWrapper) error {
	if wrapper == nil {
		return errors.New("invalid import data: missing wrapper")
	}
	if wrapper.Version == "" {
		return errors.New("invalid import data: missing version")
	}
	for _, entry := range wrapper.Words {
		if err := ValidateWordEntry(entry); err != nil {
			return err
		}
	}
	return nil
}
