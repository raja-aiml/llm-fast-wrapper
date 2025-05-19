package embeddings

import (
	"context"
	"database/sql"

	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"go.uber.org/zap"
)

// Fetcher encapsulates embedding retrieval logic across cache, pgvector, and OpenAI APIs.
type Fetcher struct {
	store VectorStore
}

// NewFetcher returns a new Fetcher using the provided VectorStore for persistence.
func NewFetcher(store VectorStore) *Fetcher {
	return &Fetcher{store: store}
}

var logger *zap.SugaredLogger

func init() {
	// Initialize a dedicated logger for embedding fetch operations
	// Write logs only to the embeddings-fetcher-error log file (no console output)
	logger = logging.InitLogger("logs/embeddings-fetcher.log")
}

// Get retrieves the embedding for the given text using a three-tier lookup:
// 1. In-memory cache
// 2. pgvector store (if configured)
// 3. OpenAI API fallback, persisting to cache and store
func (f *Fetcher) Get(ctx context.Context, text string) ([]float32, error) {
	logger.Debugf("Fetching embedding for text: %q", text)
	// 1. In-memory cache
	if cached, ok := GetCachedEmbedding(text); ok {
		logger.Debugf("Cache hit for text: %q", text)
		return cached, nil
	}
	logger.Infof("Cache miss for text: %q", text)

	// 2. Attempt to load from pgvector store, if configured
	if f.store != nil {
		logger.Debugf("Attempting to load embedding for %q from pgvector store", text)
		vec, err := f.store.Get(ctx, text)
		if err == nil {
			logger.Infof("Loaded embedding for %q from pgvector store", text)
			SetCachedEmbedding(text, vec)
			return vec, nil
		}
		if err != sql.ErrNoRows {
			logger.Errorf("Error retrieving embedding for %q from pgvector store: %v", text, err)
			return nil, err
		}
		logger.Infof("No pgvector embedding found for %q; generating via API", text)
	} else {
		logger.Debugf("No pgvector store configured; will generate embedding for %q via API", text)
	}

	// 3. Generate embedding via OpenAI API
	logger.Debugf("Generating embedding via OpenAI API for %q", text)
	vec, err := GetEmbedding(text, "")
	if err != nil {
		logger.Errorf("Failed to generate embedding via OpenAI API for %q: %v", text, err)
		return nil, err
	}
	logger.Infof("Generated embedding via OpenAI API for %q", text)

	// 4. Store in in-memory cache
	SetCachedEmbedding(text, vec)
	logger.Debugf("Stored embedding in cache for %q", text)

	// 5. Persist to pgvector store, if configured
	if f.store != nil {
		logger.Debugf("Persisting embedding for %q to pgvector store", text)
		if perr := f.store.Store(ctx, text, vec); perr != nil {
			logger.Errorf("Failed to persist embedding for %q to pgvector store: %v", text, perr)
		} else {
			logger.Infof("Successfully persisted embedding for %q to pgvector store", text)
		}
	}
	logger.Debugf("Returning embedding for %q", text)
	return vec, nil
}
