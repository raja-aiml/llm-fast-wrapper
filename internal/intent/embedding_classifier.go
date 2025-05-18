package intent

import (
   "fmt"
   "sort"

   "github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
   "github.com/raja.aiml/llm-fast-wrapper/internal/logging"
   "go.uber.org/zap"
)
// logger is the package-level logger for intent classification
var logger *zap.SugaredLogger

func init() {
   // initialize logger writing to logs/intent.log
   logger = logging.InitLogger("logs/intent.log")
}

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
   logger.Infof("Starting ClassifyIntentWithEmbeddings: query=%q, promptDir=%q, extension=%q", query, promptDir, extension)
   // Load all strategies
   strategies, paths, err := LoadStrategyFiles(promptDir, extension)
   if err != nil {
       logger.Errorf("Error loading strategy files: %v", err)
       return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, err
	}

   logger.Infof("Loaded %d strategies", len(strategies))
   // Get embedding for the query
   queryEmbedding, err := embeddings.GetEmbedding(query, "")
   if err != nil {
       logger.Errorf("Failed to get embedding for query: %v", err)
       return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, fmt.Errorf("failed to get embedding for query: %w", err)
	}

   logger.Debug("Query embedding obtained")
   // Prepare texts and names for batch embedding
   var strategyTexts []string
   var strategyNames []string

	for name, content := range strategies {
		strategyTexts = append(strategyTexts, content)
		strategyNames = append(strategyNames, name)
	}

   strategyEmbeddings := embeddings.GetEmbeddingsBatch(strategyTexts, "")
   logger.Infof("Obtained embeddings for %d strategies", len(strategyEmbeddings))

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

   // If no match was found (due to all errors), return default strategy
   if bestScore < 0 {
       logger.Warn("No valid strategy embeddings found; returning default strategy")
       return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, nil
	}

   logger.Infof("Best match: %s (score=%.4f)", bestMatch.Name, bestScore)
   return bestMatch, nil
}

// ClassifyIntentWithEmbeddingsThreshold finds the most similar prompt strategy for a given query
// using OpenAI embeddings, but only returns a match if the similarity score is above the threshold
// Otherwise returns the default strategy
func ClassifyIntentWithEmbeddingsThreshold(query, promptDir, extension string, threshold float64) (EmbeddingStrategyMatch, error) {
   logger.Infof("Applying threshold=%.4f for query", threshold)
   match, err := ClassifyIntentWithEmbeddings(query, promptDir, extension)
   if err != nil {
       logger.Errorf("Error in embedding classification: %v", err)
   }
   if err != nil || match.Score < threshold {
       logger.Infof("Threshold check failed (score=%.4f < threshold=%.4f); returning default strategy", match.Score, threshold)
       return EmbeddingStrategyMatch{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0,
			Content: DefaultStrategy,
		}, err
	}

   logger.Infof("Threshold check passed (score=%.4f >= threshold=%.4f); returning match %s", match.Score, threshold, match.Name)
   return match, nil
}

// GetTopNMatchesWithEmbeddings returns the top N matching strategies for a given query
// using OpenAI embeddings for semantic matching
func GetTopNMatchesWithEmbeddings(query, promptDir, extension string, n int) ([]EmbeddingStrategyMatch, error) {
   logger.Infof("Starting GetTopNMatchesWithEmbeddings: query=%q, promptDir=%q, ext=%q, topN=%d", query, promptDir, extension, n)
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

   logger.Infof("Loaded %d strategies", len(strategies))
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

   logger.Debug("Query embedding obtained for top-N matching")
   // Prepare texts and names for batch embedding
   var strategyTexts []string
   var strategyNames []string

	for name, content := range strategies {
		strategyTexts = append(strategyTexts, content)
		strategyNames = append(strategyNames, name)
	}

   strategyEmbeddings := embeddings.GetEmbeddingsBatch(strategyTexts, "")
   logger.Infof("Computed embeddings for %d strategies", len(strategyEmbeddings))

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
   logger.Infof("Sorted %d matches; returning top %d", len(matches), n)

	// Return top N matches (or all if less than N available)
   if len(matches) > n {
       logger.Infof("Truncating matches to top %d entries", n)
       return matches[:n], nil
   }
   return matches, nil
}
