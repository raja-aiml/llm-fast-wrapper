package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"go.uber.org/zap"
)

// PostgresStore stores embeddings in a PostgreSQL pgvector table
type PostgresStore struct {
	db        *sql.DB
	dimension int
	logger    *zap.SugaredLogger
}

// NewPostgresStore connects to PostgreSQL and ensures table & index exist
func NewPostgresStore(dsn string, dimension int) (*PostgresStore, error) {
	logger := logging.InitLogger("logs/pgvector.log")
	logger.Infof("Connecting to PostgreSQL with DSN: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	logger.Info("Connected to PostgreSQL")

	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		return nil, fmt.Errorf("create extension: %w", err)
	}
	logger.Info("pgvector extension ensured")

	// Ensure embeddings table exists with the correct vector dimension
	createTable := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS embeddings (
			text TEXT PRIMARY KEY,
			embedding vector(%d)
		);`, dimension)
	if _, err := db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("ensure embeddings table: %w", err)
	}
	logger.Infof("Ensured embeddings table with dimension=%d", dimension)

	// Ensure ivfflat index exists on the embeddings column
	indexQuery := `
		CREATE INDEX IF NOT EXISTS idx_embeddings_embedding
		ON embeddings USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100);`
	if _, err := db.Exec(indexQuery); err != nil {
		return nil, fmt.Errorf("ensure embeddings index: %w", err)
	}
	logger.Info("Ensured ivfflat index on embeddings")

	return &PostgresStore{
		db:        db,
		dimension: dimension,
		logger:    logger,
	}, nil
}

// Store stores an embedding in the PostgreSQL database
func (s *PostgresStore) Store(ctx context.Context, text string, embedding []float32) error {
	vectorLiteral := toVectorLiteral(embedding)
	s.logger.Debugf("Storing embedding for %q", text)

	query := `
		INSERT INTO embeddings (text, embedding)
		VALUES ($1, $2::vector)
		ON CONFLICT (text) DO UPDATE SET embedding = EXCLUDED.embedding;`
	_, err := s.db.ExecContext(ctx, query, text, vectorLiteral)
	if err != nil {
		s.logger.Warnf("Store failed for %q: %v", text, err)
	} else {
		s.logger.Debugf("Stored embedding successfully for %q", text)
	}
	return err
}

// Get retrieves an embedding from the PostgreSQL database
func (s *PostgresStore) Get(ctx context.Context, text string) ([]float32, error) {
	s.logger.Debugf("Retrieving embedding for %q", text)

	row := s.db.QueryRowContext(ctx, `SELECT embedding FROM embeddings WHERE text = $1`, text)
	var vec string
	if err := row.Scan(&vec); err != nil {
		s.logger.Infof("No embedding found for %q: %v", text, err)
		return nil, err
	}

	parsed, err := parseVectorLiteral(vec)
	if err != nil {
		s.logger.Errorf("Parse error for embedding %q: %v", text, err)
		return nil, err
	}
	s.logger.Debugf("Retrieved embedding for %q", text)
	return parsed, nil
}

// SearchByEmbedding searches for similar embeddings in the database
func (s *PostgresStore) SearchByEmbedding(ctx context.Context, embedding []float32, k int) ([]storage.SimilarItem, error) {
	vectorLiteral := toVectorLiteral(embedding)
	s.logger.Debugf("Searching top-%d embeddings", k)

	query := fmt.Sprintf(`
		SELECT text, embedding <=> $1::vector AS distance
		FROM embeddings
		ORDER BY embedding <=> $1::vector
		LIMIT %d`, k)
	rows, err := s.db.QueryContext(ctx, query, vectorLiteral)
	if err != nil {
		s.logger.Errorf("Search query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []storage.SimilarItem
	for rows.Next() {
		var item storage.SimilarItem
		if err := rows.Scan(&item.Text, &item.Distance); err != nil {
			s.logger.Errorf("Scan failed: %v", err)
			return nil, err
		}
		item.Similarity = 1 - item.Distance
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		s.logger.Errorf("Row iteration error: %v", err)
		return nil, err
	}

	s.logger.Infof("Found %d similar items", len(items))
	return items, nil
}

// SearchStrategies searches for similar prompt strategies in the database
func (s *PostgresStore) SearchStrategies(ctx context.Context, embedding []float32, threshold float64, maxResults int) ([]storage.StrategyItem, error) {
	vectorLiteral := toVectorLiteral(embedding)
	s.logger.Debugf("DB-based searching strategies with threshold=%.4f, maxResults=%d", threshold, maxResults)
	query := `SELECT id, name, path, content, similarity FROM find_similar_strategies($1::vector, $2, $3)`
	rows, err := s.db.QueryContext(ctx, query, vectorLiteral, threshold, maxResults)
	if err != nil {
		s.logger.Errorf("SearchStrategies query failed: %v", err)
		return nil, err
	}
	defer rows.Close()
	var items []storage.StrategyItem
	for rows.Next() {
		var it storage.StrategyItem
		if err := rows.Scan(&it.ID, &it.Name, &it.Path, &it.Content, &it.Similarity); err != nil {
			s.logger.Errorf("SearchStrategies scan failed: %v", err)
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		s.logger.Errorf("SearchStrategies rows error: %v", err)
		return nil, err
	}
	s.logger.Infof("SearchStrategies returned %d items", len(items))
	return items, nil
}

// UpsertStrategy inserts or updates a prompt strategy record in prompt_strategies table.
func (s *PostgresStore) UpsertStrategy(ctx context.Context, name, path, content string, embedding []float32) (int64, error) {
	vectorLiteral := toVectorLiteral(embedding)
	s.logger.Debugf("Upserting strategy %q into prompt_strategies (dimension=%d)", name, s.dimension)
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
		s.logger.Errorf("Failed to upsert strategy %q: %v", name, err)
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		s.logger.Errorf("Failed to get RowsAffected for strategy %q: %v", name, err)
		return 0, err
	}
	if rows > 0 {
		s.logger.Debugf("Upsert affected %d row(s) for strategy %q", rows, name)
	} else {
		s.logger.Debugf("Skipped upsert for strategy %q (no change)", name)
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

// parseVectorLiteral parses a pgvector literal string into []float32.
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
