package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/raja.aiml/llm-fast-wrapper/internal/intent"
)

func main() {
	dir := flag.String("dir", "strategies", "path to your .md strategy files")
	ext := flag.String("ext", ".md", "strategy file extension")
	threshold := flag.Float64("threshold", 0.5, "minimum similarity to accept")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("Usage: %s [options] <query>", flag.CommandLine.Name())
	}
	query := flag.Arg(0)

	match, err := intent.ClassifyIntentWithEmbeddingsThreshold(query, *dir, *ext, *threshold)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("Matched:  %s (score=%.4f)\n", match.Name, match.Score)
	fmt.Printf("Strategy: %s\n\n%s\n", match.Path, match.Content)
}

// go run ./cmd/intent --dir  ../prompting-strategies --threshold 0.6 "How do I use goroutines?"
