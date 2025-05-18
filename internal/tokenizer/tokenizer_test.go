package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Whitespace only",
			input:    "   \t\n  ",
			expected: []string{},
		},
		{
			name:     "Simple sentence",
			input:    "Hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Sentence with punctuation",
			input:    "Hello, world! How are you?",
			expected: []string{"hello", "world", "how", "are", "you"},
		},
		{
			name:     "Mixed case",
			input:    "HeLLo WoRLd",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Extra whitespace",
			input:    "  Hello   world  ",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Numbers and special characters",
			input:    "Hello123 world! 456-789",
			expected: []string{"hello123", "world", "456", "789"},
		},
		{
			name:     "Technical content",
			input:    "func main() { fmt.Println(\"Hello, world!\") }",
			expected: []string{"func", "main", "fmt", "println", "hello", "world"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SimpleTokenize(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
