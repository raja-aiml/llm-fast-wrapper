package embeddings

import (
	"context"
	"database/sql"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/api"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/cache"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"go.uber.org/zap"
)

// Service encapsulates the embedding functionality, providing a unified API
// This consolidates functionality from both the original Service and Fetcher
type Service struct {
	provider api.Provider
	store    storage.VectorStore
	cache    *cache.Cache
	logger   *zap.SugaredLogger
}

// NewService creates a new Embedding Service
func NewService(provider api.Provider, store storage.VectorStore) *Service {
	return &Service{
		provider: provider,
		store:    store,
		cache:    cache.NewCache(),
		logger:   logging.InitLogger("logs/embeddings-service.log"),
	}
}

// Get retrieves the embedding for the given text using a three-tier lookup
func (s *Service) Get(ctx context.Context, text string) ([]float32, error) {
	s.logger.Debugf("Fetching embedding for text: %q", text)

	// 1. In-memory cache
	if cached, found := s.cache.Get(text); found {
		s.logger.Debugf("Cache hit for text: %q", text)
		return cached, nil
	}
	s.logger.Infof("Cache miss for text: %q", text)

	// 2. Attempt storage fetch
	if s.store != nil {
		s.logger.Debugf("Attempting storage retrieval for text: %q", text)
		vec, err := s.store.Get(ctx, text)
		if err == nil {
			s.cache.Set(text, vec)
			s.logger.Infof("Embedding loaded from storage for %q", text)
			return vec, nil
		}
		if err != sql.ErrNoRows {
			s.logger.Warnf("Storage retrieval failed for %q: %v", text, err)
		} else {
			s.logger.Infof("No storage record found for %q", text)
		}
	}

	// 3. Fallback to provider (e.g., OpenAI)
	s.logger.Debugf("Calling embedding provider for %q", text)
	vec, err := s.provider.GenerateEmbedding(ctx, text, "")
	if err != nil {
		s.logger.Errorf("Embedding generation failed for %q: %v", text, err)
		return nil, err
	}
	s.logger.Infof("Successfully generated embedding for %q", text)

	// Cache it
	s.cache.Set(text, vec)
	s.logger.Debugf("Cached embedding in memory for %q", text)

	// Persist to store
	if s.store != nil {
		s.logger.Debugf("Persisting embedding to storage for %q", text)
		if perr := s.store.Store(ctx, text, vec); perr != nil {
			s.logger.Warnf("Could not persist embedding for %q to storage: %v", text, perr)
		} else {
			s.logger.Infof("Successfully persisted embedding for %q to storage", text)
		}
	}

	return vec, nil
}

// GetBatch generates embeddings for multiple texts
func (s *Service) GetBatch(ctx context.Context, texts []string) (map[string][]float32, error) {
	result := make(map[string][]float32)
	var uncachedTexts []string

	// Check cache first
	for _, text := range texts {
		if cached, found := s.cache.Get(text); found {
			result[text] = cached
		} else {
			uncachedTexts = append(uncachedTexts, text)
		}
	}

	// If all found in cache, return early
	if len(uncachedTexts) == 0 {
		return result, nil
	}

	// Generate remaining embeddings
	embeddings := s.provider.GenerateEmbeddingsBatch(ctx, uncachedTexts, "")

	// Process results
	for i, embResult := range embeddings {
		if embResult.Error != nil {
			s.logger.Warnf("Failed to generate embedding for text %q: %v", uncachedTexts[i], embResult.Error)
			continue
		}

		// Add to result
		text := uncachedTexts[i]
		result[text] = embResult.Embedding

		// Cache it
		s.cache.Set(text, embResult.Embedding)

		// Store it
		if s.store != nil {
			if err := s.store.Store(ctx, text, embResult.Embedding); err != nil {
				s.logger.Warnf("Failed to store embedding for %q: %v", text, err)
			}
		}
	}

	return result, nil
}

// ClearCache clears the in-memory cache
func (s *Service) ClearCache() {
	s.cache.Clear()
	s.logger.Info("Embedding cache cleared")
}

// GetCacheSize returns the number of entries in the in-memory cache
func (s *Service) GetCacheSize() int {
	return s.cache.Size()
}

// CosineSimilarity calculates the cosine similarity between two embedding vectors
// This is now just a proxy to the central implementation
func (s *Service) CosineSimilarity(vec1, vec2 []float32) float32 {
	return api.CosineSimilarity(vec1, vec2)
}
