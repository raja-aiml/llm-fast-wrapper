package internal

import (
	"context"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	"github.com/raja.aiml/llm-fast-wrapper/internal/intent"
	"go.uber.org/zap"
)

func matchQuery(
	ctx context.Context,
	embedder *embeddings.Service,
	store interface{},
	strategies map[string]string,
	query string,
	threshold float64,
	useDB bool,
	logger *zap.SugaredLogger,
) *intent.MatchResult {
	if useDB {
		//psStore := store.(*postgres.PostgresStore)
		advancedStore, ok := store.(storage.AdvancedVectorStore)
		if !ok {
			logger.Fatal("Store does not support advanced search features")
		}
		vec, err := embedder.Get(ctx, query)
		if err != nil {
			logger.Fatalf("Embedding failed for query: %v", err)
		}
		items, err := advancedStore.SearchStrategies(ctx, vec, threshold, 1)
		if err != nil {
			logger.Fatalf("DB search failed: %v", err)
		}
		if len(items) == 0 {
			return &intent.MatchResult{Name: "Default Strategy", Path: "built-in", Score: 0, Content: intent.DefaultStrategy}
		}
		item := items[0]
		return &intent.MatchResult{Name: item.Name, Path: item.Path, Score: item.Similarity, Content: item.Content}
	}
	result, err := intent.MatchBestStrategy(ctx, query, strategies, embedder, threshold)
	if err != nil {
		logger.Fatalf("Match error: %v", err)
	}
	return result
}
