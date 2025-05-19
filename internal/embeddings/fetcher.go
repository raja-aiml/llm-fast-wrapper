package embeddings

import (
	"context"
	"database/sql"
)

type Fetcher struct {
	store VectorStore
}

func NewFetcher(store VectorStore) *Fetcher {
	return &Fetcher{store: store}
}

func (f *Fetcher) Get(ctx context.Context, text string) ([]float32, error) {
	// Cache
	if cached, ok := GetCachedEmbedding(text); ok {
		return cached, nil
	}

	// PgVector
	if f.store != nil {
		vec, err := f.store.Get(ctx, text)
		if err == nil {
			SetCachedEmbedding(text, vec)
			return vec, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	// OpenAI
	vec, err := GetEmbedding(text, "")
	if err != nil {
		return nil, err
	}
	SetCachedEmbedding(text, vec)

	// Persist
	if f.store != nil {
		_ = f.store.Store(ctx, text, vec) // log but ignore error
	}
	return vec, nil
}
