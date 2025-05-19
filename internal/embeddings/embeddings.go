package embeddings

import (
	"context"
	"sync"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/api"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/cache"
)

// EmbeddingResult represents the result of an embedding operation
type EmbeddingResult struct {
	Embedding []float32
	Error     error
}

// Global variables for the package
var (
	defaultProvider api.Provider
	defaultOnce     sync.Once
	defaultInitErr  error
	globalCache     = NewCache()
)

// initProvider initializes the default provider
func initProvider() error {
	defaultOnce.Do(func() {
		var err error
		defaultProvider, err = api.NewOpenAIProvider()
		if err != nil {
			defaultInitErr = err
			return
		}
	})
	return defaultInitErr
}

// GetEmbedding retrieves or generates an embedding for the given text
func GetEmbedding(text string, modelName string) ([]float32, error) {
	if err := initProvider(); err != nil {
		return nil, err
	}

	// Check if we have a cached embedding
	cachedEmbedding, found := globalCache.Get(text)
	if found {
		return cachedEmbedding, nil
	}

	// If not in cache, generate new embedding
	ctx := context.Background()
	embedding, err := defaultProvider.GenerateEmbedding(ctx, text, modelName)
	if err != nil {
		return nil, err
	}

	// Cache the embedding
	globalCache.Set(text, embedding)
	return embedding, nil
}

// GetEmbeddingsBatch gets embeddings for multiple texts
func GetEmbeddingsBatch(texts []string, modelName string) []EmbeddingResult {
	if err := initProvider(); err != nil {
		results := make([]EmbeddingResult, len(texts))
		for i := range results {
			results[i].Error = err
		}
		return results
	}

	ctx := context.Background()
	results := make([]EmbeddingResult, len(texts))

	// Check cache first
	uncachedTexts := make([]string, 0)
	uncachedIndices := make([]int, 0)

	for i, text := range texts {
		if cached, found := globalCache.Get(text); found {
			results[i].Embedding = cached
		} else {
			uncachedTexts = append(uncachedTexts, text)
			uncachedIndices = append(uncachedIndices, i)
		}
	}

	// Generate embeddings for uncached texts
	if len(uncachedTexts) > 0 {
		batchResults := defaultProvider.GenerateEmbeddingsBatch(ctx, uncachedTexts, modelName)

		// Map results back to the original indices
		for i, result := range batchResults {
			idx := uncachedIndices[i]
			results[idx].Embedding = result.Embedding
			results[idx].Error = result.Error

			// Cache successful embeddings
			if result.Error == nil {
				globalCache.Set(uncachedTexts[i], result.Embedding)
			}
		}
	}

	return results
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(vec1, vec2 []float32) float32 {
	return api.CosineSimilarity(vec1, vec2)
}

// ClearCache clears the embedding cache
func ClearCache() {
	globalCache.Clear()
}

// NewCache creates a new embedding cache
func NewCache() *cache.Cache {
	return cache.NewCache()
}
