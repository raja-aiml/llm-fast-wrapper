package intent

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/raja.aiml/llm-fast-wrapper/internal/tokenizer"
)

// DefaultStrategy is a minimal built-in fallback strategy when no files are found
const DefaultStrategy = `# Default Strategy

This is a minimal default strategy that can be used when no other strategies are available.
It provides basic guidance for general queries and should be enhanced with domain-specific strategies.

## General Guidelines
- Understand the query context
- Identify key requirements
- Provide concise and relevant responses
- Ask clarifying questions when necessary
`

// StrategyMatch represents a matched strategy with its similarity score
type StrategyMatch struct {
	Name    string  // Strategy name (derived from filename)
	Path    string  // Path to the strategy file
	Score   float64 // Similarity score (0-1)
	Content string  // Content of the strategy file
}

// LoadPromptStrategies recursively reads all files with the specified extension from a directory
// and returns their contents combined into a single string, with each file's
// basename (without extension) used as a heading.
// If the directory doesn't exist or contains no matching files, it returns a default strategy.
func LoadPromptStrategies(rootDir string, extension string) string {
	// Set default extension if not provided
	if extension == "" {
		extension = ".md"
	}

	// Ensure extension starts with a dot
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	var strategies []string
	var filePaths []string

	// Check if directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		// Directory doesn't exist, use default strategy
		return DefaultStrategy
	}

	// Walk the directory recursively
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Only process files with matching extension
		if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(extension)) {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		// If walking fails, return default strategy
		return DefaultStrategy
	}

	// If no matching files found, return default strategy
	if len(filePaths) == 0 {
		return DefaultStrategy
	}

	// Sort file paths to ensure consistent ordering
	sort.Strings(filePaths)

	// Process each file
	for _, path := range filePaths {
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			strategies = append(strategies, fmt.Sprintf("❌ Failed to read %s: %v", path, err))
			continue
		}

		// Extract filename without extension to use as heading
		baseName := filepath.Base(path)
		headingName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		headingName = strings.ReplaceAll(headingName, "_", " ")
		headingName = strings.Title(headingName) // Convert to title case

		// Add the heading and file content
		strategies = append(strategies, fmt.Sprintf("# %s\n\n%s", headingName, string(content)))
	}

	// Join all strategies with separators
	return strings.Join(strategies, "\n\n---\n\n")
}

// LoadStrategyFiles recursively reads all files with the specified extension from a directory
// and returns a map of strategy names to their file contents.
// This is used for classification rather than aggregation.
func LoadStrategyFiles(rootDir string, extension string) (map[string]string, map[string]string, error) {
	// Set default extension if not provided
	if extension == "" {
		extension = ".md"
	}

	// Ensure extension starts with a dot
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	strategyContents := make(map[string]string)
	strategyPaths := make(map[string]string)

	// Check if directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		// Add default strategy
		strategyContents["Default Strategy"] = DefaultStrategy
		strategyPaths["Default Strategy"] = "built-in"
		return strategyContents, strategyPaths, nil
	}

	// Walk the directory recursively
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Only process files with matching extension
		if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(extension)) {
			content, err := os.ReadFile(filepath.Clean(path))
			if err != nil {
				return nil // Skip files we can't read
			}

			// Extract filename without extension to use as strategy name
			baseName := filepath.Base(path)
			strategyName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
			strategyName = strings.ReplaceAll(strategyName, "_", " ")
			strategyName = strings.Title(strategyName) // Convert to title case

			strategyContents[strategyName] = string(content)
			strategyPaths[strategyName] = path
		}
		return nil
	})

	if err != nil || len(strategyContents) == 0 {
		// Add default strategy if directory walk failed or no files found
		strategyContents["Default Strategy"] = DefaultStrategy
		strategyPaths["Default Strategy"] = "built-in"
	}

	return strategyContents, strategyPaths, nil
}

// CosineSimilarity calculates the cosine similarity between two vectors
func CosineSimilarity(vec1, vec2 map[string]float64) float64 {
	var dotProduct, norm1, norm2 float64

	// Calculate dot product
	for term, weight1 := range vec1 {
		if weight2, exists := vec2[term]; exists {
			dotProduct += weight1 * weight2
		}
	}

	// Calculate vector norms
	for _, weight := range vec1 {
		norm1 += weight * weight
	}
	for _, weight := range vec2 {
		norm2 += weight * weight
	}

	// Avoid division by zero
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// Tokenize splits text into tokens and creates a term frequency map
func Tokenize(text string) map[string]float64 {
	// Convert to lowercase and tokenize
	tokens := tokenizer.SimpleTokenize(text)

	// Create term frequency map
	termFreq := make(map[string]float64)
	for _, token := range tokens {
		termFreq[token]++
	}

	return termFreq
}

// ClassifyIntent finds the most similar prompt strategy for a given query
// Returns the best match along with its similarity score
func ClassifyIntent(query, promptDir, extension string) (StrategyMatch, error) {
	// Load all strategies
	strategies, paths, err := LoadStrategyFiles(promptDir, extension)
	if err != nil {
		return StrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, err
	}

	// Tokenize query
	queryVector := Tokenize(query)

	// Find the most similar strategy
	var bestMatch StrategyMatch
	bestScore := -1.0

	for name, content := range strategies {
		strategyVector := Tokenize(content)
		similarity := CosineSimilarity(queryVector, strategyVector)

		if similarity > bestScore {
			bestScore = similarity
			bestMatch = StrategyMatch{
				Name:    name,
				Path:    paths[name],
				Score:   similarity,
				Content: content,
			}
		}
	}

	return bestMatch, nil
}

// ClassifyIntentWithThreshold finds the most similar prompt strategy for a given query
// but only returns a match if the similarity score is above the threshold
// Otherwise returns the default strategy
func ClassifyIntentWithThreshold(query, promptDir, extension string, threshold float64) (StrategyMatch, error) {
	match, err := ClassifyIntent(query, promptDir, extension)
	if err != nil || match.Score < threshold {
		return StrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, err
	}

	return match, nil
}

// GetTopNMatches returns the top N matching strategies for a given query
func GetTopNMatches(query, promptDir, extension string, n int) ([]StrategyMatch, error) {
	// Load all strategies
	strategies, paths, err := LoadStrategyFiles(promptDir, extension)
	if err != nil {
		return []StrategyMatch{{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}}, err
	}

	// Tokenize query
	queryVector := Tokenize(query)

	// Calculate similarity scores for all strategies
	var matches []StrategyMatch
	for name, content := range strategies {
		strategyVector := Tokenize(content)
		similarity := CosineSimilarity(queryVector, strategyVector)

		matches = append(matches, StrategyMatch{
			Name:    name,
			Path:    paths[name],
			Score:   similarity,
			Content: content,
		})
	}

	// Sort by similarity score (descending)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Return top N matches (or all if less than N available)
	if len(matches) > n {
		return matches[:n], nil
	}
	return matches, nil
}
