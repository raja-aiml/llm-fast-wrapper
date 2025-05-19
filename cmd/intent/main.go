package main

import (
   "context"
   "database/sql"
   "flag"
   "fmt"
   "sort"

   "github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
   "github.com/raja.aiml/llm-fast-wrapper/internal/intent"
   "github.com/raja.aiml/llm-fast-wrapper/internal/logging"
)

func main() {
   // Command-line flags
   dir := flag.String("dir", "strategies", "path to your .md strategy files")
   ext := flag.String("ext", ".md", "strategy file extension")
   threshold := flag.Float64("threshold", 0.5, "minimum similarity to accept")
   dbDSN := flag.String("db-dsn", "", "Postgres DSN for pgvector store (optional)")
   dbDim := flag.Int("db-dim", 1536, "Vector dimension for pgvector store")
   flag.Parse()

   // Initialize structured logging
   logger := logging.InitLogger("logs/intent.log")
   defer func() { _ = logger.Sync() }()

   if flag.NArg() < 1 {
       logger.Fatalf("Usage: %s [options] <query>", flag.CommandLine.Name())
   }
   query := flag.Arg(0)

   // 1. Load strategy files
   strategies, paths, err := intent.LoadStrategyFiles(*dir, *ext)
   if err != nil {
       logger.Fatalf("Error loading strategy files: %v", err)
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

   // Initialize optional pgvector store
   var store *embeddings.PostgresStore
   if *dbDSN != "" {
       store, err = embeddings.NewPostgresStore(*dbDSN, *dbDim)
       if err != nil {
           logger.Fatalf("Failed to initialize pgvector store: %v", err)
       }
       logger.Infof("Connected to pgvector store, DSN=%s", *dbDSN)
   }

   // 2. Generate query embedding (memory cache + API)
   ctx := context.Background()
   queryEmb, err := embeddings.GetEmbedding(query, "")
   if err != nil {
       logger.Fatalf("Failed to get embedding for query: %v", err)
   }
   logger.Infof("Query embedding generated")

   // 3-5. Retrieve or generate embeddings per strategy (memory cache, pgvector, API), then compute similarity
   bestScore := -1.0
   var bestName string
   for _, name := range names {
       text := strategies[name]
       var emb []float32
       // Check in-memory cache
       if cached, ok := embeddings.GetCachedEmbedding(text); ok {
           logger.Debugf("Cache hit for strategy %s", name)
           emb = cached
       } else if store != nil {
           // Attempt to load from pgvector
           pgEmb, err := store.Get(ctx, text)
           if err == nil {
               logger.Infof("Loaded embedding for %s from pgvector store", name)
               embeddings.SetCachedEmbedding(text, pgEmb)
               emb = pgEmb
           } else if err == sql.ErrNoRows {
               logger.Infof("No pgvector embedding for %s; generating via API", name)
               emb, err = embeddings.GetEmbedding(text, "")
               if err != nil {
                   logger.Errorf("Failed to generate embedding for %s: %v", name, err)
                   continue
               }
               // Persist to pgvector
               if perr := store.Store(ctx, text, emb); perr != nil {
                   logger.Errorf("Failed to store embedding for %s in pgvector: %v", name, perr)
               } else {
                   logger.Infof("Stored embedding for %s in pgvector store", name)
               }
           } else {
               logger.Errorf("Error retrieving embedding for %s from pgvector: %v", name, err)
               continue
           }
       } else {
           // No pgvector; generate via API
           logger.Debugf("Generating embedding for %s via API", name)
           emb, err = embeddings.GetEmbedding(text, "")
           if err != nil {
               logger.Errorf("Failed to generate embedding for %s: %v", name, err)
               continue
           }
       }
       // Compute similarity
       score := float64(embeddings.CosineSimilarity(queryEmb, emb))
       logger.Debugf("Strategy %s similarity score=%.4f", name, score)
       if score > bestScore {
           bestScore = score
           bestName = name
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
