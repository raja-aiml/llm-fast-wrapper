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
	seedOnly := flag.Bool("seed-only", false, "seed prompt_strategies table and exit (no matching)")
	useDB := flag.Bool("use-db", false, "use database-based matching via prompt_strategies table")
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

	// Prepare context
	ctx := context.Background()

   // If using PostgresStore, seed prompt_strategies (and optionally exit)
   if psStore, ok := store.(*embeddings.PostgresStore); ok && psStore != nil {
       var seeded, skipped int64
       for name, content := range strategies {
           path := paths[name]
           // Compute or retrieve embedding
           vec, err := embedder.Get(ctx, content)
           if err != nil {
               logger.Errorf("Failed to compute embedding for strategy %q: %v", name, err)
               continue
           }
           // Upsert into prompt_strategies
           affected, err := psStore.UpsertStrategy(ctx, name, path, content, vec)
           if err != nil {
               logger.Errorf("Failed to upsert strategy %q: %v", name, err)
           } else if affected > 0 {
               seeded++
               logger.Debugf("Strategy %q seeded/updated", name)
           } else {
               skipped++
           }
       }
       logger.Infof("Seeding complete: %d inserted/updated, %d skipped", seeded, skipped)
       // Exit early if in seed-only mode
       if *seedOnly {
           return
       }
   }

   // Match: either via DB-backed search or in-memory file-based matching
   var result *intent.MatchResult
   if *useDB {
       // Database-based matching using prompt_strategies table
       psStore := store.(*embeddings.PostgresStore)
       // Compute query embedding
       qEmb, err := embedder.Get(ctx, query)
       if err != nil {
           logger.Fatalf("Failed to compute embedding for query: %v", err)
       }
       // Search for top strategy
       items, err := psStore.SearchStrategies(ctx, qEmb, *threshold, 1)
       if err != nil {
           logger.Fatalf("DB search failed: %v", err)
       }
       if len(items) == 0 {
           result = &intent.MatchResult{Name: "Default Strategy", Path: "built-in", Score: 0.0, Content: intent.DefaultStrategy}
       } else {
           item := items[0]
           result = &intent.MatchResult{Name: item.Name, Path: item.Path, Score: item.Similarity, Content: item.Content}
       }
   } else {
       // In-memory file-based matching
       result, err = intent.MatchBestStrategy(ctx, query, strategies, embedder, *threshold)
       if err != nil {
           logger.Fatalf("Match failed: %v", err)
       }
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

//  go run ./cmd/intent \
//           --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
//           --db-dim 1536 \
//           --dir ../prompting-strategies \
//           --threshold 0.6 \
//           "Explain TCP"
