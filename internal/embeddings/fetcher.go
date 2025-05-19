package embeddings

import (
	"context"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"go.uber.org/zap"
)

// Fetcher encapsulates embedding retrieval logic across cache, pgvector, and OpenAI APIs
type Fetcher struct {
	store  storage.VectorStore
	logger *zap.SugaredLogger
}

// NewFetcher returns a new Fetcher using the provided VectorStore for persistence
func NewFetcher(store storage.VectorStore) *Fetcher {
	return &Fetcher{
		store:  store,
		logger: logging.InitLogger("logs/embeddings-fetcher.log"),
	}
}

// Get retrieves the embedding for the given text using a three-tier lookup
func (f *Fetcher) Get(ctx context.Context, text string) ([]float32, error) {
	f.logger.Debugf("Fetching embedding for text: %q", text)

	// Check if we have a cached embedding
	cachedEmbedding, found := globalCache.Get(text)
	if found {
		f.logger.Debugf("Cache hit for text: %q", text)
		return cachedEmbedding, nil
	}
	f.logger.Infof("Cache miss for text: %q", text)

	// Check storage if provided
	if f.store != nil {
		f.logger.Debugf("Attempting storage retrieval for text: %q", text)
		vec, err := f.store.Get(ctx, text)
		if err == nil {
			// Cache and return
			globalCache.Set(text, vec)
			f.logger.Infof("Embedding loaded from storage for %q", text)
			return vec, nil
		}
	}

	// Generate from provider
	f.logger.Debugf("Generating embedding via API for: %q", text)
	vec, err := GetEmbedding(text, "")
	if err != nil {
		f.logger.Errorf("Embedding generation failed for %q: %v", text, err)
		return nil, err
	}

	// Store in persistent storage if available
	if f.store != nil {
		f.logger.Debugf("Persisting embedding for %q", text)
		if err := f.store.Store(ctx, text, vec); err != nil {
			f.logger.Warnf("Failed to persist embedding: %v", err)
		}
	}

	return vec, nil
}
