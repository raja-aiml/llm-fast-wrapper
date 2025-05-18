package tokenizer

import (
	"regexp"
	"strings"
)

var (
	// Regular expression to match non-alphanumeric characters
	nonAlphaNumRegex = regexp.MustCompile(`[^a-zA-Z0-9\s]`)

	// Regular expression to match multiple whitespace characters
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// SimpleTokenize takes a string and returns a slice of lowercase tokens
// with punctuation removed and normalized whitespace
func SimpleTokenize(text string) []string {
	// Skip empty text
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove non-alphanumeric characters
	text = nonAlphaNumRegex.ReplaceAllString(text, " ")

	// Normalize whitespace
	text = whitespaceRegex.ReplaceAllString(text, " ")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	// Split into tokens
	tokens := strings.Split(text, " ")

	// Remove empty tokens
	var result []string
	for _, token := range tokens {
		if token != "" {
			result = append(result, token)
		}
	}

	return result
}
