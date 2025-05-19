package storage

import "context"

// VectorStore defines the interface for external embedding stores
type VectorStore interface {
	// Get retrieves an embedding vector for the given text
	Get(ctx context.Context, text string) ([]float32, error)

	// Store saves an embedding vector for the given text
	Store(ctx context.Context, text string, vec []float32) error
}

// SimilarItem represents a result from similarity search
type SimilarItem struct {
	Text       string
	Distance   float32
	Similarity float32
}

// StrategyItem represents a prompt strategy record returned by DB search
type StrategyItem struct {
	ID         int
	Name       string
	Path       string
	Content    string
	Similarity float64
}

// AdvancedVectorStore adds search capabilities to the basic VectorStore interface
type AdvancedVectorStore interface {
	VectorStore

	// SearchByEmbedding finds the most similar vectors to the given embedding
	SearchByEmbedding(ctx context.Context, embedding []float32, k int) ([]SimilarItem, error)

	// SearchStrategies searches for similar prompt strategies
	SearchStrategies(ctx context.Context, embedding []float32, threshold float64, maxResults int) ([]StrategyItem, error)

	// UpsertStrategy inserts or updates a prompt strategy record
	UpsertStrategy(ctx context.Context, name, path, content string, embedding []float32) (int64, error)
}
