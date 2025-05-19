# Intent Matching CLI

This CLI tool loads prompt strategies from `.md` files, computes their embeddings (via cache, pgvector, or OpenAI), and finds the most relevant match for a given query.

## 🧠 How It Works

1. Loads all `.md` files from the specified strategy directory.
2. Computes or retrieves their embeddings:
   - Checks in-memory cache
   - Optionally uses PostgreSQL pgvector if `--db-dsn` is set
   - Falls back to OpenAI Embedding API
3. If pgvector is enabled, it upserts the data into the `prompt_strategies` table.
4. Computes the embedding for the user query.
5. Matches the most similar strategy based on cosine similarity.

---

## 📦 Usage

```bash
go run ./cmd/intent \
  --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
  --db-dim 1536 \
  --dir ../prompting-strategies \
  --threshold 0.6 \
  "Explain TCP"
```

## 🔧 Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--dir` | Path to directory containing strategy files | `"strategies"` |
| `--ext` | File extension for strategy files | `".md"` |
| `--threshold` | Minimum similarity score to accept a match | `0.5` |
| `--db-dsn` | PostgreSQL connection string for pgvector (optional) | `""` |
| `--db-dim` | Vector dimension for pgvector embeddings | `1536` |


When using the `--db-dsn` flag:
1. The tool connects to a PostgreSQL database with pgvector extension
2. Strategy embeddings are stored for faster retrieval
3. The `prompt_strategies` table must exist with appropriate schema

```bash
go run cmd/migrate/main.go drop \
             --db-dsn "postgresql://user:pass@host:5432/db?sslmode=disable"

go run cmd/migrate/main.go migrate \
                 --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
                 --db-dim 1536

go run cmd/intent/main.go \
          --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
          --db-dim 1536 \
          --dir ../prompting-strategies \
          --ext .md \
          --seed-only

go run cmd/intent/main.go \
                 --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
                 --db-dim 1536 \
                 --use-db \
                 --dir ../prompting-strategies/ \
                 --ext .md \
                 "Your query here"

```