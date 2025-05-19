package internal

import (
	"context"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage/postgres"
	"go.uber.org/zap"
)

func maybeSeed(ctx context.Context, embedder *embeddings.Service, store storage.VectorStore, strategies map[string]string, paths map[string]string, logger *zap.SugaredLogger) int64 {
	psStore, ok := store.(*postgres.PostgresStore)
	if !ok || psStore == nil {
		return 0
	}

	var seeded int64
	for name, content := range strategies {
		path := paths[name]
		vec, err := embedder.Get(ctx, content)
		if err != nil {
			logger.Errorf("Embedding failed for %q: %v", name, err)
			continue
		}
		if affected, err := psStore.UpsertStrategy(ctx, name, path, content, vec); err != nil {
			logger.Errorf("Upsert failed for %q: %v", name, err)
		} else if affected > 0 {
			logger.Infof("Strategy %q seeded", name)
			seeded++
		}
	}
	logger.Infof("Seeding complete: %d strategies inserted/updated", seeded)
	return seeded
}
