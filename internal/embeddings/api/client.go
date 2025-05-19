package api

import "context"

// EmbeddingResult represents the result of an embedding operation
type EmbeddingResult struct {
	Embedding []float32
	Error     error
}

// Provider defines the interface for embedding generation services
type Provider interface {
	// GenerateEmbedding generates a single embedding for the given text
	GenerateEmbedding(ctx context.Context, text string, modelName string) ([]float32, error)

	// GenerateEmbeddingsBatch generates embeddings for multiple texts in parallel
	GenerateEmbeddingsBatch(ctx context.Context, texts []string, modelName string) []EmbeddingResult

	// SetDefaultModel changes the default embedding model
	SetDefaultModel(model string)
}
