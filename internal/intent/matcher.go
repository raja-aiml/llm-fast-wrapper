package intent

import (
	"context"
	"sort"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
)

type MatchResult struct {
	Name    string
	Path    string
	Score   float64
	Content string
}

func MatchBestStrategy(
	ctx context.Context,
	query string,
	strategies map[string]string,
	embedder *embeddings.Service,
	threshold float64,
) (*MatchResult, error) {
	queryEmb, err := embeddings.GetEmbedding(query, "")
	if err != nil {
		return nil, err
	}

	names := sortedKeys(strategies)
	bestScore := -1.0
	var bestName string

	for _, name := range names {
		text := strategies[name]
		emb, err := embedder.Get(ctx, text)
		if err != nil {
			continue
		}
		score := float64(embeddings.CosineSimilarity(queryEmb, emb))
		if score > bestScore {
			bestScore = score
			bestName = name
		}
	}

	if bestScore < threshold {
		return &MatchResult{
			Name:    "Default Strategy",
			Path:    "built-in",
			Score:   0.0,
			Content: DefaultStrategy,
		}, nil
	}

	return &MatchResult{
		Name:    bestName,
		Path:    "", // filled by caller using paths[bestName]
		Score:   bestScore,
		Content: strategies[bestName],
	}, nil
}

func sortedKeys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
