package internal

import (
	"context"
	"os"

	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/api"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	pgstore "github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage/postgres"
	"github.com/raja.aiml/llm-fast-wrapper/internal/intent"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runIntentCommand is the main command logic
func runIntentCommand(cmd *cobra.Command, args []string) {
	logger := logging.InitLogger("logs/intent.log", "stdout")
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()
	cfg.Query = parseQuery(cfg, args, logger)

	strategies, paths := loadStrategies(cfg, logger)
	store := connectStore(cfg, logger)
	embedder := initEmbedder(cfg, store, logger)

	handleMatchOrSeed(ctx, cfg, embedder, store, strategies, paths, logger)
}

// --- Helper functions ---

func parseQuery(cfg *Config, args []string, logger *zap.SugaredLogger) string {
	if cfg.SeedOnly {
		return ""
	}
	if len(args) == 0 {
		logger.Fatalf("Missing query. Usage: intent [flags] <query>")
	}
	return args[0]
}

func loadStrategies(cfg *Config, logger *zap.SugaredLogger) (map[string]string, map[string]string) {
	strategies, paths, err := intent.LoadStrategyFiles(cfg.Dir, cfg.Ext)
	if err != nil {
		logger.Fatalf("Failed to load strategy files: %v", err)
	}
	if len(strategies) == 0 {
		printDefaultStrategy()
		os.Exit(0)
	}
	logger.Infof("Loaded %d strategies from %s", len(strategies), cfg.Dir)
	return strategies, paths
}

func connectStore(cfg *Config, logger *zap.SugaredLogger) storage.VectorStore {
	if cfg.DbDSN == "" {
		return nil
	}
	store, err := pgstore.NewPostgresStore(cfg.DbDSN, cfg.DbDim)
	if err != nil {
		logger.Fatalf("Failed to connect to pgvector store: %v", err)
	}
	logger.Infof("Connected to pgvector store: %s", cfg.DbDSN)
	return store
}

func initEmbedder(cfg *Config, store storage.VectorStore, logger *zap.SugaredLogger) *embeddings.Service {
	// log the configuration
	logger.Infof("Using DB  with DSN: %s", cfg.DbDSN)
	provider, err := api.NewOpenAIProvider()
	if err != nil {
		logger.Fatalf("Failed to initialize OpenAI provider: %v", err)
	}
	logger.Info("Initialized OpenAI provider")
	return embeddings.NewService(provider, store)
}

func handleMatchOrSeed(
	ctx context.Context,
	cfg *Config,
	embedder *embeddings.Service,
	store storage.VectorStore,
	strategies map[string]string,
	paths map[string]string,
	logger *zap.SugaredLogger,
) {
	if seeded := maybeSeed(ctx, embedder, store, strategies, paths, logger); cfg.SeedOnly {
		logger.Infof("Seed-only mode complete: %d strategies processed", seeded)
		return
	}

	result := matchQuery(ctx, embedder, store, strategies, cfg.Query, cfg.Threshold, cfg.UseDB, logger)

	if path, ok := paths[result.Name]; ok {
		result.Path = path
	}
	printMatch(result)
}
