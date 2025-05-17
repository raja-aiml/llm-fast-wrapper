package llm

import (
	"fmt"
	"time"
)

type OpenAIStreamer struct{}

func NewOpenAIStreamer() Streamer { return &OpenAIStreamer{} }

func (o *OpenAIStreamer) Stream(prompt string) (<-chan string, error) {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for i := 0; i < 10; i++ {
			ch <- fmt.Sprintf("token-%d", i)
			time.Sleep(200 * time.Millisecond)
		}
	}()
	return ch, nil
}
