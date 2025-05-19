package llm

import (
	"strings"
	"time"
)

type OpenAIStreamer struct{}

func NewOpenAIStreamer() Streamer { return &OpenAIStreamer{} }

// Stream returns mock chat completion chunks that follow the OpenAI streaming
// specification. The implementation simply splits the prompt into tokens and
// emits one token per chunk with slight delays to mimic network latency.
func (o *OpenAIStreamer) Stream(prompt string) (<-chan ChatCompletionChunk, error) {
	ch := make(chan ChatCompletionChunk)
	go func() {
		defer close(ch)
		tokens := strings.Fields(prompt)
		id := "chatcmpl-mock"
		created := time.Now().Unix()
		for _, t := range tokens {
			ch <- ChatCompletionChunk{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: created,
				Choices: []ChatCompletionChoice{{
					Delta: Delta{Content: t + " "},
					Index: 0,
				}},
			}
			time.Sleep(100 * time.Millisecond)
		}
		stop := "stop"
		ch <- ChatCompletionChunk{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Choices: []ChatCompletionChoice{{
				Delta:        Delta{},
				Index:        0,
				FinishReason: &stop,
			}},
		}
	}()
	return ch, nil
}
