package help

import (
   "testing"

   "github.com/stretchr/testify/assert"
)

func TestLoadUsageMarkdownInvalidPath(t *testing.T) {
   out := LoadUsageMarkdown("nonexistent.md")
   assert.Contains(t, out, "❌ Failed to load help markdown:")
}

func TestLoadUsageMarkdownValidPath(t *testing.T) {
   // usage.md is present in this directory
   out := LoadUsageMarkdown("usage.md")
   assert.NotEmpty(t, out)
   assert.NotContains(t, out, "❌ Failed to load help markdown:")
}