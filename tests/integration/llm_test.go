//go:build integration
// +build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/raja.aiml/llm-fast-wrapper/internal/llm"
)

func TestOpenAIStreamer(t *testing.T) {
	streamer := llm.NewOpenAIStreamer()
	ch, err := streamer.Stream("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	timeout := time.After(3 * time.Second)
	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				if count == 0 {
					t.Fatalf("no chunks received")
				}
				return
			}
			count++
			if chunk.Choices[0].FinishReason != nil {
				if *chunk.Choices[0].FinishReason != "stop" {
					t.Fatalf("unexpected finish reason: %v", *chunk.Choices[0].FinishReason)
				}
			}
		case <-timeout:
			t.Fatal("timeout")
		}
	}
}
