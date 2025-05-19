package api

import (
	"context"
	"errors"
	"os"
	"sync"

	openai "github.com/openai/openai-go"
	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
)

// OpenAIProvider is an implementation of Provider using OpenAI's API
type OpenAIProvider struct {
	client       *openai.Client
	defaultModel string
	mutex        sync.RWMutex
}

var (
	defaultOpenAIModel = openai.EmbeddingModelTextEmbeddingAda002
	initOnce           sync.Once
	initErr            error
	defaultClient      *openai.Client
)

// NewOpenAIProvider creates a new OpenAI embedding provider
func NewOpenAIProvider() (*OpenAIProvider, error) {
	if err := initializeClient(); err != nil {
		return nil, err
	}

	return &OpenAIProvider{
		client:       defaultClient,
		defaultModel: defaultOpenAIModel,
	}, nil
}

// initializeClient initializes the OpenAI client
func initializeClient() error {
	initOnce.Do(func() {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			initErr = errors.New("OPENAI_API_KEY environment variable is not set")
			return
		}

		// Create the OpenAI client using internal config
		c := config.NewClient(apiKey, "")
		defaultClient = &c
	})
	return initErr
}

// SetDefaultModel changes the default embedding model
func (p *OpenAIProvider) SetDefaultModel(model string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.defaultModel = model
}

// GenerateEmbedding retrieves or generates an embedding for the given text
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string, modelName string) ([]float32, error) {
	// Use default model if none specified
	p.mutex.RLock()
	if modelName == "" {
		modelName = p.defaultModel
	}
	p.mutex.RUnlock()

	// Prepare request parameters for embedding
	params := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
		Model: modelName,
	}

	// Call the OpenAI Embedding API
	res, err := p.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, errors.New("no embedding returned from API")
	}

	// Convert embedding from float64 to float32
	raw := res.Data[0].Embedding
	embedding := make([]float32, len(raw))
	for i, v := range raw {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// GenerateEmbeddingsBatch gets embeddings for multiple texts in parallel
func (p *OpenAIProvider) GenerateEmbeddingsBatch(ctx context.Context, texts []string, modelName string) []EmbeddingResult {
	results := make([]EmbeddingResult, len(texts))

	// Use default model if none specified
	p.mutex.RLock()
	if modelName == "" {
		modelName = p.defaultModel
	}
	p.mutex.RUnlock()

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
				embedding, err := p.GenerateEmbedding(ctx, text, modelName)
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
