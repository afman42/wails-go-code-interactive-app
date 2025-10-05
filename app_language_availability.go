//go:build windows || linux

package main

import (
	"os/exec"
	"strings"
)

type LanguageAvailability struct {
	System  []string `json:"system"`
	Bundled []string `json:"bundled"`
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func (a *App) ListLanguageAvailability(languages []string) LanguageAvailability {
	availability := LanguageAvailability{}

	system := make([]string, 0, len(languages))
	bundled := make([]string, 0, len(languages))

	for _, language := range languages {
		trimmed := strings.TrimSpace(language)
		if trimmed == "" {
			continue
		}

		if path, err := exec.LookPath(trimmed); err == nil && path != "" {
			system = append(system, trimmed)
		}

		if a.hasBundledExecutable(trimmed) {
			bundled = append(bundled, trimmed)
		}
	}

	availability.System = dedupeStrings(system)
	availability.Bundled = dedupeStrings(bundled)
	return availability
}
