package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/api"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
)

// MemoryStore implements an in-memory vector store
type MemoryStore struct {
	embeddings map[string][]float32
	mutex      sync.RWMutex
}

// NewMemoryStore creates a new in-memory vector store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		embeddings: make(map[string][]float32),
	}
}

// Get retrieves an embedding from the in-memory store
func (s *MemoryStore) Get(ctx context.Context, text string) ([]float32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	embedding, ok := s.embeddings[text]
	if !ok {
		return nil, fmt.Errorf("no embedding found for text: %s", text)
	}

	return embedding, nil
}

// Store saves an embedding to the in-memory store
func (s *MemoryStore) Store(ctx context.Context, text string, vec []float32) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.embeddings[text] = vec
	return nil
}

// SearchByEmbedding finds similar vectors in the memory store
func (s *MemoryStore) SearchByEmbedding(ctx context.Context, embedding []float32, k int) ([]storage.SimilarItem, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var items []storage.SimilarItem

	// Calculate similarity with all stored embeddings
	for text, vec := range s.embeddings {
		similarity := api.CosineSimilarity(embedding, vec)
		items = append(items, storage.SimilarItem{
			Text:       text,
			Distance:   1 - similarity,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Similarity > items[j].Similarity
	})

	// Return top k results
	if len(items) > k {
		items = items[:k]
	}

	return items, nil
}
