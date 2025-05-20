package llm

import (
	"context"
	"strings"
	"time"
)

type OpenAIStreamer struct{}

func NewOpenAIStreamer() Streamer { return &OpenAIStreamer{} }

// Stream returns mock chat completion chunks that follow the OpenAI streaming
// specification. The implementation simply splits the prompt into tokens and
// emits one token per chunk with slight delays to mimic network latency.
func (o *OpenAIStreamer) Stream(ctx context.Context, prompt string) (<-chan ChatCompletionChunk, error) {
	ch := make(chan ChatCompletionChunk)
	go func() {
		defer close(ch)
		tokens := strings.Fields(prompt)
		id := "chatcmpl-mock"
		created := time.Now().Unix()
		for _, t := range tokens {
			select {
			case <-ctx.Done():
				return
			default:
			}

			select {
			case ch <- ChatCompletionChunk{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: created,
				Choices: []ChatCompletionChoice{{
					Delta: Delta{Content: t + " "},
					Index: 0,
				}},
			}:
			case <-ctx.Done():
				return
			}

			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}
		stop := "stop"
		select {
		case ch <- ChatCompletionChunk{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Choices: []ChatCompletionChoice{{
				Delta:        Delta{},
				Index:        0,
				FinishReason: &stop,
			}},
		}:
		case <-ctx.Done():
			return
		}
	}()
	return ch, nil
}
