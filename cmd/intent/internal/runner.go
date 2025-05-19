package internal

import (
	"context"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/api"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage/postgres"
	"github.com/raja.aiml/llm-fast-wrapper/internal/intent"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"github.com/spf13/cobra"
)

func runIntentCommand(cmd *cobra.Command, args []string) {
	logger := logging.InitLogger("logs/intent.log", "stdout")
	defer func() { _ = logger.Sync() }()

	var query string
	if !seedOnly {
		if len(args) < 1 {
			logger.Fatalf("Missing query. Usage: %s [flags] <query>", cmd.Name())
		}
		query = args[0]
	}

	// Load strategy files
	strategies, paths, err := intent.LoadStrategyFiles(dir, ext)
	if err != nil {
		logger.Fatalf("LoadStrategyFiles failed: %v", err)
	}
	if len(strategies) == 0 {
		printDefaultStrategy()
		return
	}

	// Optional pgvector
	var store storage.VectorStore
	if dbDSN != "" {
		store, err = postgres.NewPostgresStore(dbDSN, dbDim)
		if err != nil {
			logger.Fatalf("Failed to connect to pgvector: %v", err)
		}
	}

	// Embedder service
	provider, err := api.NewOpenAIProvider()
	if err != nil {
		logger.Fatalf("OpenAI init error: %v", err)
	}
	service := embeddings.NewService(provider, store)
	ctx := context.Background()

	// Seeding mode
	if seeded := maybeSeed(ctx, service, store, strategies, paths, logger); seedOnly {
		logger.Infof("Seed-only mode complete: %d strategies processed", seeded)
		return
	}

	// Match mode
	result := matchQuery(ctx, service, store, strategies, query, threshold, useDB, logger)

	// Output
	if path, ok := paths[result.Name]; ok {
		result.Path = path
	}
	printMatch(result)
}
