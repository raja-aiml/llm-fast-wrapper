# 🚀 Intent Matching CLI

This CLI classifies user queries by matching them to the most relevant prompt strategy using vector similarity (OpenAI embeddings + optional pgvector DB).

---

## 🧠 How It Works

1. Loads all `.md` files from the configured strategy directory.
2. Computes or retrieves embeddings for each strategy:
   - Uses in-memory cache
   - Optionally fetches/stores from a PostgreSQL `pgvector` store
   - Falls back to OpenAI for new embeddings
3. For a given query:
   - Embeds the query
   - Finds the best match from strategies
   - Returns the best strategy above the configured similarity threshold

---

## 🧪 Caching Behavior

- The in-memory cache resets on every run.
- The Postgres store is persistent and prevents unnecessary OpenAI calls across runs.
- After a one-time seed (`--seed-only`), all strategy embeddings are reused.

---

## ⚙️ Usage

### 🔄 Migrate DB Schema
```bash
go run cmd/migrate/main.go migrate \
  --db-dsn "postgresql://user:pass@localhost:5432/llmlogs?sslmode=disable" \
  --db-dim 1536
```

### 🌱 Seed Strategy Embeddings
```bash
go run cmd/intent/main.go \
  --db-dsn "postgresql://user:pass@localhost:5432/llmlogs?sslmode=disable" \
  --db-dim 1536 \
  --dir ./prompting-strategies \
  --seed-only
```

### 🔍 Match Query with DB-Based Search
```bash
go run cmd/intent/main.go \
  --db-dsn "postgresql://user:pass@localhost:5432/llmlogs?sslmode=disable" \
  --db-dim 1536 \
  --use-db \
  --dir ./prompting-strategies \
  "Explain TCP"
```

### 🧮 Match In-Memory (no DB)
```bash
go run cmd/intent/main.go \
  --dir ./prompting-strategies \
  "Explain TCP"
```

### 🔧 CLI Flags


### 📁 Source Structure

```text
cmd/intent
├── internal/
│   ├── flags.go       # Cobra flags binding
│   ├── root.go        # Cobra command and Execute()
│   ├── runner.go      # Main command logic dispatcher
│   ├── matcher.go     # Query matching logic
│   ├── seeder.go      # Seeder for strategy embeddings
│   └── output.go      # Match result printing
├── main.go            # Thin entrypoint
├── README.md
```


### ✅ Design Principles
* Separation of Concerns: CLI, embedding, matching, and DB logic are decoupled
* KISS: Logic is split into small, focused functions
* YAGNI: Only supports essential flags and paths
* DRY: Reuses service abstractions (embeddings.Service, postgres.PostgresStore)
* SOLID: Uses interfaces for embedding stores and batch matchers


