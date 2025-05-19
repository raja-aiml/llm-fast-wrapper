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

	// Use only for dev: Drop existing table (make this configurable in future)
	pgLogger.Warn("Dropping existing embeddings table (development mode only)")
	if _, err := db.Exec(`DROP TABLE IF EXISTS embeddings`); err != nil {
		return nil, fmt.Errorf("drop embeddings: %w", err)
	}

	createTable := fmt.Sprintf(`
		CREATE TABLE embeddings (
			text TEXT PRIMARY KEY,
			embedding vector(%d)
		);`, dimension)
	if _, err := db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	pgLogger.Infof("Created embeddings table with dimension=%d", dimension)

	indexQuery := `
		CREATE INDEX IF NOT EXISTS idx_embeddings_embedding
		ON embeddings USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100);`
	if _, err := db.Exec(indexQuery); err != nil {
		return nil, fmt.Errorf("create index: %w", err)
	}
	pgLogger.Info("Created ivfflat index on embeddings")

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

// UpsertStrategy inserts or updates a prompt strategy record in prompt_strategies table.
func (s *PostgresStore) UpsertStrategy(ctx context.Context, name, path, content string, embedding []float32) error {
   vectorLiteral := toVectorLiteral(embedding)
   pgLogger.Debugf("Upserting strategy %q into prompt_strategies (dimension=%d)", name, s.dimension)
   // Only update if content or path has changed
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
   _, err := s.db.ExecContext(ctx, upsert, name, path, content, vectorLiteral)
   if err != nil {
       pgLogger.Errorf("Failed to upsert strategy %q: %v", name, err)
   } else {
       pgLogger.Debugf("Successfully upserted strategy %q", name)
   }
   return err
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
