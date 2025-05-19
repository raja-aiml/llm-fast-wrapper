package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
	"github.com/raja.aiml/llm-fast-wrapper/internal/intent"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
)

func main() {
	// CLI flags
	dir := flag.String("dir", "strategies", "path to your .md strategy files")
	ext := flag.String("ext", ".md", "strategy file extension")
	threshold := flag.Float64("threshold", 0.5, "minimum similarity to accept")
	dbDSN := flag.String("db-dsn", "", "Postgres DSN for pgvector store (optional)")
	dbDim := flag.Int("db-dim", 1536, "Vector dimension for pgvector store")
	flag.Parse()

	// Logger
	logger := logging.InitLogger("logs/intent.log")
	defer func() { _ = logger.Sync() }()

	// Validate args
	if flag.NArg() < 1 {
		logger.Fatalf("Usage: %s [options] <query>", flag.CommandLine.Name())
	}
	query := flag.Arg(0)

	// Load strategies
	strategies, paths, err := intent.LoadStrategyFiles(*dir, *ext)
	if err != nil {
		logger.Fatalf("Error loading strategy files: %v", err)
	}
	if len(strategies) == 0 {
		fmt.Println("Matched:  Default Strategy (score=0.0000)")
		fmt.Printf("Strategy: built-in\n\n" + intent.DefaultStrategy)
		return
	}
	logger.Infof("Loaded %d strategies from dir=%s ext=%s", len(strategies), *dir, *ext)

	// Optional: connect pgvector
	var store embeddings.VectorStore
	if *dbDSN != "" {
		store, err = embeddings.NewPostgresStore(*dbDSN, *dbDim)
		if err != nil {
			logger.Fatalf("Failed to connect to pgvector: %v", err)
		}
		logger.Infof("Connected to pgvector store: DSN=%s", *dbDSN)
	}

	// Fetcher = cache → pgvector → OpenAI
	embedder := embeddings.NewFetcher(store)

	// Match
	ctx := context.Background()
	result, err := intent.MatchBestStrategy(ctx, query, strategies, embedder, *threshold)
	if err != nil {
		logger.Fatalf("Match failed: %v", err)
	}

	// Use `paths` map to get file path
	if path, ok := paths[result.Name]; ok {
		result.Path = path
	} else {
		result.Path = "unknown"
	}

	// Output
	fmt.Printf("Matched:  %s (score=%.4f)\n", result.Name, result.Score)
	fmt.Printf("Strategy: %s\n\n%s\n", result.Path, result.Content)
	logger.Infof("Final match: %s (score=%.4f) path=%s", result.Name, result.Score, result.Path)
}
