package embeddings

import (
	"context"
	"errors"
	"math"
	"os"
	"sync"
	"time"

	openai "github.com/openai/openai-go"
	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/cache"
)

// Global variables for the package
var (
	client       *openai.Client
	defaultModel = openai.EmbeddingModelTextEmbeddingAda002
	initOnce     sync.Once
	initErr      error
	// Create a global cache instance
	globalCache = cache.NewCache()
)

// EmbeddingResult represents the result of an embedding operation
type EmbeddingResult struct {
	Embedding []float32
	Error     error
}

// initClient initializes the OpenAI client
func initClient() error {
	initOnce.Do(func() {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			initErr = errors.New("OPENAI_API_KEY environment variable is not set")
			return
		}

		// Create the OpenAI client using internal config
		c := config.NewClient(apiKey, "")
		client = &c
	})
	return initErr
}

// GetEmbedding retrieves or generates an embedding for the given text
func GetEmbedding(text string, modelName string) ([]float32, error) {
	if err := initClient(); err != nil {
		return nil, err
	}

	// Check if we have a cached embedding
	cachedEmbedding, found := globalCache.Get(text)
	if found {
		return cachedEmbedding, nil
	}

	// If not in cache, generate new embedding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use default model if none specified
	if modelName == "" {
		modelName = defaultModel
	}

	// Prepare request parameters for embedding
	params := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
		Model: modelName,
	}
	// Call the OpenAI Embedding API
	res, err := client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return nil, errors.New("no embedding returned from API")
	}
	// Convert embedding from float64 to float32 and cache it
	raw := res.Data[0].Embedding
	embedding := make([]float32, len(raw))
	for i, v := range raw {
		embedding[i] = float32(v)
	}
	globalCache.Set(text, embedding)
	return embedding, nil
}

// GetEmbeddingsBatch gets embeddings for multiple texts in parallel
func GetEmbeddingsBatch(texts []string, modelName string) []EmbeddingResult {
	results := make([]EmbeddingResult, len(texts))

	// Check for initialization error
	if err := initClient(); err != nil {
		for i := range results {
			results[i].Error = err
		}
		return results
	}

	// Process in batches to avoid rate limits
	const batchSize = 20
	for start := 0; start < len(texts); start += batchSize {
		end := start + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		// Get batch slice
		batch := texts[start:end]
		batchResults := make([]EmbeddingResult, len(batch))

		// Process batch in parallel
		var wg sync.WaitGroup
		for i, text := range batch {
			wg.Add(1)
			go func(i int, text string) {
				defer wg.Done()
				embedding, err := GetEmbedding(text, modelName)
				batchResults[i] = EmbeddingResult{
					Embedding: embedding,
					Error:     err,
				}
			}(i, text)
		}
		wg.Wait()

		// Copy batch results to main results
		for i, result := range batchResults {
			results[start+i] = result
		}
	}

	return results
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(vec1, vec2 []float32) float32 {
	if len(vec1) == 0 || len(vec2) == 0 || len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, norm1, norm2 float32

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	// Avoid division by zero
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	// Compute cosine similarity: dot(vec1,vec2)/(||vec1||*||vec2||)
	return dotProduct / (float32(math.Sqrt(float64(norm1))) * float32(math.Sqrt(float64(norm2))))
}

// SetDefaultModel changes the default embedding model
func SetDefaultModel(model string) {
	defaultModel = model
}

// ClearCache clears the embedding cache
func ClearCache() {
	globalCache.Clear()
}

// GetCachedEmbedding returns the embedding for the given text from the in-memory cache.
// The boolean indicates whether the embedding was found.
func GetCachedEmbedding(text string) ([]float32, bool) {
	return globalCache.Get(text)
}

// SetCachedEmbedding stores the embedding for the given text in the in-memory cache.
func SetCachedEmbedding(text string, embedding []float32) {
	globalCache.Set(text, embedding)
}
