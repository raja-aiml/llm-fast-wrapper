package llm_test

import (
    "testing"
    "time"

    "github.com/raja.aiml/llm-fast-wrapper/internal/llm"
)

func TestOpenAIStreamer(t *testing.T) {
    streamer := llm.NewOpenAIStreamer()
    ch, err := streamer.Stream("hi")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    count := 0
    timeout := time.After(3 * time.Second)
    for {
        select {
        case _, ok := <-ch:
            if !ok {
                if count != 10 {
                    t.Fatalf("expected 10 tokens, got %d", count)
                }
                return
            }
            count++
        case <-timeout:
            t.Fatal("timeout")
        }
    }
}
