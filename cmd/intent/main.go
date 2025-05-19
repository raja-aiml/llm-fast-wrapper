package main

import (
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
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

	// 1. Load strategy files
	strategies, paths, err := intent.LoadStrategyFiles(*dir, *ext)
	if err != nil {
		log.Fatalf("Error loading strategy files: %v", err)
	}
	if len(strategies) == 0 {
		// No strategies found; use default
		fmt.Printf("Matched:  Default Strategy (score=0.0000)\n")
		fmt.Printf("Strategy: built-in\n\n%s\n", intent.DefaultStrategy)
		return
	}

	// Sort strategy names for deterministic order
	var names []string
	for name := range strategies {
		names = append(names, name)
	}
	sort.Strings(names)

	// 2. Generate query embedding
	queryEmb, err := embeddings.GetEmbedding(query, "")
	if err != nil {
		log.Fatalf("Failed to get embedding for query: %v", err)
	}

	// 3. Batch-embed strategies
	var texts []string
	for _, name := range names {
		texts = append(texts, strategies[name])
	}
	results := embeddings.GetEmbeddingsBatch(texts, "")

	// 4. Compute cosine similarities and 5. apply threshold logic
	bestScore := -1.0
	var bestName string
	for i, res := range results {
		if res.Error != nil {
			continue
		}
		score := float64(embeddings.CosineSimilarity(queryEmb, res.Embedding))
		if score > bestScore {
			bestScore = score
			bestName = names[i]
		}
	}

	// 6. Determine final match
	var matchName, matchPath, matchContent string
	var matchScore float64
	if bestScore < *threshold || bestScore < 0 {
		matchName = "Default Strategy"
		matchPath = "built-in"
		matchScore = 0.0
		matchContent = intent.DefaultStrategy
	} else {
		matchName = bestName
		matchPath = paths[bestName]
		matchScore = bestScore
		matchContent = strategies[bestName]
	}

	// Output the match
	fmt.Printf("Matched:  %s (score=%.4f)\n", matchName, matchScore)
	fmt.Printf("Strategy: %s\n\n%s\n", matchPath, matchContent)
}

// go run ./cmd/intent --dir  ../prompting-strategies --threshold 0.6 "How do I use goroutines?"
