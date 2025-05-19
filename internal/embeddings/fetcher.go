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
	logger = logging.InitLogger("logs/embeddings-fetcher.log")
	logger.Debug("Initialized embedding fetcher logger")
}

// Get retrieves the embedding for the given text using a three-tier lookup
func (f *Fetcher) Get(ctx context.Context, text string) ([]float32, error) {
	logger.Debugf("Fetching embedding for text: %q", text)

	// 1. In-memory cache
	if cached, ok := GetCachedEmbedding(text); ok {
		logger.Debugf("Cache hit for text: %q", text)
		return cached, nil
	}
	logger.Infof("Cache miss for text: %q", text)

	// 2. Attempt pgvector store fetch
	if f.store != nil {
		logger.Debugf("Attempting pgvector retrieval for text: %q", text)
		vec, err := f.store.Get(ctx, text)
		if err == nil {
			SetCachedEmbedding(text, vec)
			logger.Infof("Embedding loaded from pgvector for %q", text)
			return vec, nil
		}
		if err != sql.ErrNoRows {
			logger.Warnf("pgvector retrieval failed for %q: %v", text, err)
		} else {
			logger.Infof("No pgvector record found for %q", text)
		}
	}

	// 3. Fallback to OpenAI embedding
	logger.Debugf("Calling OpenAI API for embedding generation: %q", text)
	vec, err := GetEmbedding(text, "")
	if err != nil {
		logger.Errorf("OpenAI embedding failed for %q: %v", text, err)
		return nil, err
	}
	logger.Infof("Successfully generated embedding for %q", text)

	// Cache it
	SetCachedEmbedding(text, vec)
	logger.Debugf("Cached embedding in memory for %q", text)

	// Persist to store
	if f.store != nil {
		logger.Debugf("Persisting embedding to pgvector for %q", text)
		if perr := f.store.Store(ctx, text, vec); perr != nil {
			logger.Warnf("Could not persist embedding for %q to pgvector store: %v", text, perr)
		} else {
			logger.Infof("Successfully persisted embedding for %q to pgvector", text)
		}
	}

	return vec, nil
}
