package models

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// NormalizeSponsorTags keeps free-form tags consistent for exact directory filters.
func NormalizeSponsorTags(values []string) ([]string, error) {
	tags := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		tag := strings.ToLower(strings.TrimSpace(value))
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > 40 {
			return nil, errors.New("tag is too long")
		}
		if !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}
	if len(tags) == 0 || len(tags) > 10 {
		return nil, errors.New("between 1 and 10 tags are required")
	}
	return tags, nil
}
