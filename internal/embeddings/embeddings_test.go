package embeddings

import (
   "os"
   "sync"
   "testing"

   "github.com/stretchr/testify/assert"
)

// TestCosineSimilarity tests the cosine similarity calculation
func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		vec1     []float32
		vec2     []float32
		expected float32
	}{
		{
			name:     "identical vectors",
			vec1:     []float32{1, 2, 3},
			vec2:     []float32{1, 2, 3},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			vec1:     []float32{1, 0, 0},
			vec2:     []float32{0, 1, 0},
			expected: 0.0,
		},
       {
           name:     "similar vectors",
           vec1:     []float32{1, 2, 3},
           vec2:     []float32{2, 1, 3},
           expected: 0.9286,
       },
		{
			name:     "empty vector 1",
			vec1:     []float32{},
			vec2:     []float32{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "empty vector 2",
			vec1:     []float32{1, 2, 3},
			vec2:     []float32{},
			expected: 0.0,
		},
		{
			name:     "different length vectors",
			vec1:     []float32{1, 2, 3},
			vec2:     []float32{1, 2},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.vec1, tt.vec2)
			if tt.expected == 0 {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.InDelta(t, float64(tt.expected), float64(result), 0.001)
			}
		})
	}
}

// TestGetEmbedding tests the embedding generation
// This test will be skipped if OPENAI_API_KEY is not set
func TestGetEmbedding(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Reset any previous state
	ClearCache()

	// Test getting an embedding
	text := "This is a test sentence for embedding."
	embedding, err := GetEmbedding(text, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, embedding)
	assert.Greater(t, len(embedding), 0)

	// Test cache hit (should not make another API call)
	embedding2, err := GetEmbedding(text, "")
	assert.NoError(t, err)
	assert.Equal(t, embedding, embedding2)
}

// TestGetEmbeddingsBatch tests batch embedding generation
// This test will be skipped if OPENAI_API_KEY is not set
func TestGetEmbeddingsBatch(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Reset any previous state
	ClearCache()

	// Test batch processing
	texts := []string{
		"This is the first test sentence.",
		"This is the second test sentence.",
		"This is the third test sentence.",
	}

	results := GetEmbeddingsBatch(texts, "")
	assert.Equal(t, len(texts), len(results))

   for _, result := range results {
		assert.NoError(t, result.Error)
		assert.NotEmpty(t, result.Embedding)
		assert.Greater(t, len(result.Embedding), 0)
	}

	// Test semantic similarity
	// Similar sentences should have higher similarity than dissimilar ones
	embedding1 := results[0].Embedding
	embedding2 := results[1].Embedding
	embedding3, err := GetEmbedding("The weather is nice today.", "")
	assert.NoError(t, err)

	// Similar sentences should have higher similarity
	sim12 := CosineSimilarity(embedding1, embedding2)
	sim13 := CosineSimilarity(embedding1, embedding3)
	assert.Greater(t, float64(sim12), float64(sim13), "Similar sentences should have higher similarity score")
}

// TestMissingAPIKey tests error handling when API key is missing
func TestMissingAPIKey(t *testing.T) {
	// Save the current API key
	origKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	// Unset the API key
	os.Setenv("OPENAI_API_KEY", "")
	
	// Reset client to force reinitialization
   client = nil
   initOnce = sync.Once{}

	// Test error handling
	_, err := GetEmbedding("test", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}
