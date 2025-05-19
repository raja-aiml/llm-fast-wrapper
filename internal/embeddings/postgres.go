package embeddings

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"go.uber.org/zap"
)

// SimilarItem represents a result from similarity search
type SimilarItem struct {
	Text     string
	Distance float32
}

// PostgresStore stores embeddings in a PostgreSQL pgvector table
type PostgresStore struct {
	db        *sql.DB
	dimension int
}

var pgLogger *zap.SugaredLogger

func init() {
	pgLogger = logging.InitLogger("logs/pgvector.log")
	pgLogger.Debug("Initialized pgvector logger")
}

// NewPostgresStore connects to PostgreSQL and ensures table & index exist
func NewPostgresStore(dsn string, dimension int) (*PostgresStore, error) {
	pgLogger.Infof("Connecting to PostgreSQL with DSN: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	pgLogger.Info("Connected to PostgreSQL")

	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		return nil, fmt.Errorf("create extension: %w", err)
	}
	pgLogger.Info("pgvector extension ensured")

   // Ensure embeddings table exists with the correct vector dimension
   createTable := fmt.Sprintf(
       `CREATE TABLE IF NOT EXISTS embeddings (
           text TEXT PRIMARY KEY,
           embedding vector(%d)
       );`, dimension)
   if _, err := db.Exec(createTable); err != nil {
       return nil, fmt.Errorf("ensure embeddings table: %w", err)
   }
   pgLogger.Infof("Ensured embeddings table with dimension=%d", dimension)

   // Ensure ivfflat index exists on the embeddings column
   indexQuery := `
       CREATE INDEX IF NOT EXISTS idx_embeddings_embedding
       ON embeddings USING ivfflat (embedding vector_cosine_ops)
       WITH (lists = 100);`
   if _, err := db.Exec(indexQuery); err != nil {
       return nil, fmt.Errorf("ensure embeddings index: %w", err)
   }
   pgLogger.Info("Ensured ivfflat index on embeddings")

	return &PostgresStore{db: db, dimension: dimension}, nil
}

func (s *PostgresStore) Store(ctx context.Context, text string, embedding []float32) error {
	vectorLiteral := toVectorLiteral(embedding)
	pgLogger.Debugf("Storing embedding for %q", text)

	query := `
		INSERT INTO embeddings (text, embedding)
		VALUES ($1, $2::vector)
		ON CONFLICT (text) DO UPDATE SET embedding = EXCLUDED.embedding;`
	_, err := s.db.ExecContext(ctx, query, text, vectorLiteral)
	if err != nil {
		pgLogger.Warnf("Store failed for %q: %v", text, err)
	} else {
		pgLogger.Debugf("Stored embedding successfully for %q", text)
	}
	return err
}

func (s *PostgresStore) Get(ctx context.Context, text string) ([]float32, error) {
	pgLogger.Debugf("Retrieving embedding for %q", text)

	row := s.db.QueryRowContext(ctx, `SELECT embedding FROM embeddings WHERE text = $1`, text)
	var vec string
	if err := row.Scan(&vec); err != nil {
		pgLogger.Infof("No embedding found for %q: %v", text, err)
		return nil, err
	}

	parsed, err := parseVectorLiteral(vec)
	if err != nil {
		pgLogger.Errorf("Parse error for embedding %q: %v", text, err)
		return nil, err
	}
	pgLogger.Debugf("Retrieved embedding for %q", text)
	return parsed, nil
}

func (s *PostgresStore) SearchByEmbedding(ctx context.Context, embedding []float32, k int) ([]SimilarItem, error) {
	vectorLiteral := toVectorLiteral(embedding)
	pgLogger.Debugf("Searching top-%d embeddings", k)

	query := fmt.Sprintf(`
		SELECT text, embedding <=> $1::vector AS distance
		FROM embeddings
		ORDER BY embedding <=> $1::vector
		LIMIT %d`, k)
	rows, err := s.db.QueryContext(ctx, query, vectorLiteral)
	if err != nil {
		pgLogger.Errorf("Search query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []SimilarItem
	for rows.Next() {
		var item SimilarItem
		if err := rows.Scan(&item.Text, &item.Distance); err != nil {
			pgLogger.Errorf("Scan failed: %v", err)
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		pgLogger.Errorf("Row iteration error: %v", err)
		return nil, err
	}

	pgLogger.Infof("Found %d similar items", len(items))
	return items, nil
}

// StrategyItem represents a prompt strategy record returned by DB search
type StrategyItem struct {
   Name       string
   Path       string
   Content    string
   Similarity float64
}

// SearchStrategies uses the find_similar_strategies SQL function to retrieve strategies
// with similarity above threshold, limited to maxResults.
func (s *PostgresStore) SearchStrategies(ctx context.Context, embedding []float32, threshold float64, maxResults int) ([]StrategyItem, error) {
   vectorLiteral := toVectorLiteral(embedding)
   pgLogger.Debugf("DB-based searching strategies with threshold=%.4f, maxResults=%d", threshold, maxResults)
   query := `SELECT name, path, content, similarity FROM find_similar_strategies($1::vector, $2, $3)`
   rows, err := s.db.QueryContext(ctx, query, vectorLiteral, threshold, maxResults)
   if err != nil {
       pgLogger.Errorf("SearchStrategies query failed: %v", err)
       return nil, err
   }
   defer rows.Close()
   var items []StrategyItem
   for rows.Next() {
       var it StrategyItem
       if err := rows.Scan(&it.Name, &it.Path, &it.Content, &it.Similarity); err != nil {
           pgLogger.Errorf("SearchStrategies scan failed: %v", err)
           return nil, err
       }
       items = append(items, it)
   }
   if err := rows.Err(); err != nil {
       pgLogger.Errorf("SearchStrategies rows error: %v", err)
       return nil, err
   }
   pgLogger.Infof("SearchStrategies returned %d items", len(items))
   return items, nil
}

// UpsertStrategy inserts or updates a prompt strategy record in prompt_strategies table.
// Returns the number of rows affected (0 = skipped, 1 = inserted or updated) and an error if any.
func (s *PostgresStore) UpsertStrategy(ctx context.Context, name, path, content string, embedding []float32) (int64, error) {
   vectorLiteral := toVectorLiteral(embedding)
   pgLogger.Debugf("Upserting strategy %q into prompt_strategies (dimension=%d)", name, s.dimension)
   upsert := `
INSERT INTO prompt_strategies (name, path, content, embedding)
VALUES ($1, $2, $3, $4::vector)
ON CONFLICT (name) DO UPDATE SET
    path = EXCLUDED.path,
    content = EXCLUDED.content,
    embedding = EXCLUDED.embedding
  WHERE prompt_strategies.content IS DISTINCT FROM EXCLUDED.content
     OR prompt_strategies.path IS DISTINCT FROM EXCLUDED.path
`
   res, err := s.db.ExecContext(ctx, upsert, name, path, content, vectorLiteral)
   if err != nil {
       pgLogger.Errorf("Failed to upsert strategy %q: %v", name, err)
       return 0, err
   }
   rows, err := res.RowsAffected()
   if err != nil {
       pgLogger.Errorf("Failed to get RowsAffected for strategy %q: %v", name, err)
       return 0, err
   }
   if rows > 0 {
       pgLogger.Debugf("Upsert affected %d row(s) for strategy %q", rows, name)
   } else {
       pgLogger.Debugf("Skipped upsert for strategy %q (no change)", name)
   }
   return rows, nil
}

// toVectorLiteral formats a []float32 as a pgvector literal "[x1,x2,...]".
func toVectorLiteral(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

func parseVectorLiteral(lit string) ([]float32, error) {
	s := strings.Trim(lit, "[]")
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	vec := make([]float32, len(parts))
	for i, p := range parts {
		var f float64
		if _, err := fmt.Sscan(p, &f); err != nil {
			return nil, err
		}
		vec[i] = float32(f)
	}
	return vec, nil
}
