## Cache Persistence and Avoiding Unnecessary OpenAI Calls

The key thing to understand is that the in-memory cache in the Fetcher only lives for the life of
the process. If you start a new `go run…` it will always be empty, and you’ll get a “cache miss” in
memory, even if you seeded before. What keeps things from calling OpenAI on every new run is the
PostgreSQL backing store. Here’s the pattern you should follow:

1. **Run the migrations** (only once, or whenever you bump your vector dim):

    ```bash
    go run cmd/migrate/main.go migrate \
      --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
      --db-dim 1536
    ```

   This will create (or re-create) the `embeddings` and `prompt_strategies` tables (and the
   pgvector extension).
2. **Seed your strategies** into the DB:

    ```bash
    go run cmd/intent/main.go \
      --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
      --db-dim 1536 \
      --dir ../prompting-strategies/ \
      --ext .md \
      --seed-only
    ```

   On this first pass, every strategy text is missing from the in-memory cache *and* the
   `embeddings` table, so you’ll see real cache misses → OpenAI calls → then each embedding is both
   written into your local process’s cache *and* upserted into the `embeddings` table (and into
   `prompt_strategies`).
3. **Any subsequent run** (whether you’re doing “real” matching, or even reseeding) **must** use
   the *same* `--db-dsn` and `--db-dim`. Then the fetcher will see:

   1. In-memory cache miss
   2. **DB store Get** hit
   3. return that vector and re-populate your in-memory cache
   4. **no** OpenAI call

   Example of doing a live match via pgvector:

    ```bash
    go run cmd/intent/main.go \
      --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
      --db-dim 1536 \
      --use-db \
      --dir ../prompting-strategies/ \
      --ext .md \
      "Explain TCP"
    ```

   Or, if you leave off `--use-db`, you still won’t hit OpenAI again for any of your strategy
   files – they’ll be loaded from the `embeddings` table – but the in-process matching loop will do the
   cosine compares in memory rather than via the SQL function.

---

**Bottom line:**

- The in-memory cache always starts empty on each `go run`.
- Persisted cache = the Postgres pgvector store.
- Always invoke the CLI with `--db-dsn` (and matching `--db-dim`) after you’ve run the migrations and
  the one-time `--seed-only` pass, and you’ll never see another cache miss (i.e. another call to OpenAI)
  for any of those same texts.
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
                 "Explain TCP"

```