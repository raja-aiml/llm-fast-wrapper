package intent

import (
	"fmt"
	"sort"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
)

// EmbeddingStrategyMatch represents a matched strategy using embeddings
type EmbeddingStrategyMatch struct {
	Name    string  // Strategy name (derived from filename)
	Path    string  // Path to the strategy file
	Score   float64 // Similarity score (0-1)
	Content string  // Content of the strategy file
}

// ClassifyIntentWithEmbeddings uses OpenAI embeddings for more accurate intent classification
// Returns the best match along with its similarity score
func ClassifyIntentWithEmbeddings(query, promptDir, extension string) (EmbeddingStrategyMatch, error) {
	// Load all strategies
	strategies, paths, err := LoadStrategyFiles(promptDir, extension)
	if err != nil {
		return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, err
	}

	// Get embedding for the query
	queryEmbedding, err := embeddings.GetEmbedding(query, "")
	if err != nil {
		return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, fmt.Errorf("failed to get embedding for query: %w", err)
	}

	// Get embeddings for all strategies in batch for efficiency
	var strategyTexts []string
	var strategyNames []string

	for name, content := range strategies {
		strategyTexts = append(strategyTexts, content)
		strategyNames = append(strategyNames, name)
	}

	strategyEmbeddings := embeddings.GetEmbeddingsBatch(strategyTexts, "")

	// Find the most similar strategy
	var bestMatch EmbeddingStrategyMatch
	bestScore := -1.0

	for i, result := range strategyEmbeddings {
		if result.Error != nil {
			continue // Skip strategies with embedding errors
		}

		name := strategyNames[i]
		content := strategyTexts[i]

		// Calculate similarity
		similarity := float64(embeddings.CosineSimilarity(queryEmbedding, result.Embedding))

		if similarity > bestScore {
			bestScore = similarity
			bestMatch = EmbeddingStrategyMatch{
				Name:    name,
				Path:    paths[name],
				Score:   similarity,
				Content: content,
			}
		}
	}

	// If no match was found (due to errors), return default strategy
	if bestScore < 0 {
		return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, nil
	}

	return bestMatch, nil
}

// ClassifyIntentWithEmbeddingsThreshold finds the most similar prompt strategy for a given query
// using OpenAI embeddings, but only returns a match if the similarity score is above the threshold
// Otherwise returns the default strategy
func ClassifyIntentWithEmbeddingsThreshold(query, promptDir, extension string, threshold float64) (EmbeddingStrategyMatch, error) {
	match, err := ClassifyIntentWithEmbeddings(query, promptDir, extension)
	if err != nil || match.Score < threshold {
		return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, err
	}

	return match, nil
}

// GetTopNMatchesWithEmbeddings returns the top N matching strategies for a given query
// using OpenAI embeddings for semantic matching
func GetTopNMatchesWithEmbeddings(query, promptDir, extension string, n int) ([]EmbeddingStrategyMatch, error) {
	// Load all strategies
	strategies, paths, err := LoadStrategyFiles(promptDir, extension)
	if err != nil {
		return []EmbeddingStrategyMatch{{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}}, err
	}

	// Get embedding for the query
	queryEmbedding, err := embeddings.GetEmbedding(query, "")
	if err != nil {
		return []EmbeddingStrategyMatch{{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}}, fmt.Errorf("failed to get embedding for query: %w", err)
	}

	// Get embeddings for all strategies in batch for efficiency
	var strategyTexts []string
	var strategyNames []string

	for name, content := range strategies {
		strategyTexts = append(strategyTexts, content)
		strategyNames = append(strategyNames, name)
	}

	strategyEmbeddings := embeddings.GetEmbeddingsBatch(strategyTexts, "")

	// Calculate similarity scores for all strategies
	var matches []EmbeddingStrategyMatch
	for i, result := range strategyEmbeddings {
		if result.Error != nil {
			continue // Skip strategies with embedding errors
		}

		name := strategyNames[i]
		content := strategyTexts[i]

		// Calculate similarity
		similarity := float64(embeddings.CosineSimilarity(queryEmbedding, result.Embedding))

		matches = append(matches, EmbeddingStrategyMatch{
			Name:    name,
			Path:    paths[name],
			Score:   similarity,
			Content: content,
		})
	}

	// If no matches were found (due to errors), return default strategy
	if len(matches) == 0 {
		return []EmbeddingStrategyMatch{{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}}, nil
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
