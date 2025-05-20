package llm

import "context"

// ChatCompletionChunk represents a single chunk of a streamed chat completion
// response matching the OpenAI specification.
type ChatCompletionChunk struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Choices []ChatCompletionChoice `json:"choices"`
}

// ChatCompletionChoice contains the partial message delta for a streamed chunk.
type ChatCompletionChoice struct {
	Delta        Delta   `json:"delta"`
	Index        int     `json:"index"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

// Delta holds the incremental content for the chunk.
type Delta struct {
	Content string `json:"content,omitempty"`
}

// Streamer streams chat completions following the OpenAI streaming format.
type Streamer interface {
	Stream(ctx context.Context, prompt string) (<-chan ChatCompletionChunk, error)
}
