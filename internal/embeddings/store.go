package embeddings

import "context"

// VectorStore defines the interface for external embedding stores (e.g., pgvector)
type VectorStore interface {
	Get(ctx context.Context, text string) ([]float32, error)
	Store(ctx context.Context, text string, vec []float32) error
}
