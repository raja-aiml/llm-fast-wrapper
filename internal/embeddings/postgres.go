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

// SimilarItem represents a single result from similarity search.
type SimilarItem struct {
    Text     string  // stored text
    Distance float32 // distance to query embedding (lower is more similar)
}

// PostgresStore stores embeddings in a PostgreSQL database using pgvector.
type PostgresStore struct {
    db        *sql.DB
    dimension int
}

var pgLogger *zap.SugaredLogger

func init() {
    // Dedicated logger for pgvector operations
    pgLogger = logging.InitLogger("logs/pgvector.log")
    defer pgLogger.Sync()
    pgLogger.Debug("Initialized pgvector logger")
}

// NewPostgresStore opens a PostgreSQL connection, ensures pgvector extension,
// creates embeddings table and an ivfflat index for similarity search.
func NewPostgresStore(dsn string, dimension int) (*PostgresStore, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    // Enable pgvector extension
    if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
        return nil, err
    }
    // Create embeddings table
    createTable := fmt.Sprintf(
        `CREATE TABLE IF NOT EXISTS embeddings (
            text TEXT PRIMARY KEY,
            embedding vector(%d)
        )`, dimension)
    if _, err := db.Exec(createTable); err != nil {
        return nil, err
    }
    // Create ivfflat index on embedding column (requires pgvector 0.4+)
    if _, err := db.Exec(
        `CREATE INDEX IF NOT EXISTS idx_embeddings_embedding
         ON embeddings USING ivfflat (embedding vector_cosine_ops)
         WITH (lists = 100)`); err != nil {
        return nil, err
    }
    return &PostgresStore{db: db, dimension: dimension}, nil
}

// Store upserts the embedding for the given text.
func (s *PostgresStore) Store(ctx context.Context, text string, embedding []float32) error {
    vectorLiteral := toVectorLiteral(embedding)
    pgLogger.Debugf("Upserting embedding for text %q into pgvector (dimension=%d)", text, s.dimension)
    upsert := `
INSERT INTO embeddings (text, embedding)
VALUES ($1, $2::vector)
ON CONFLICT (text) DO UPDATE SET embedding = EXCLUDED.embedding
`
    _, err := s.db.ExecContext(ctx, upsert, text, vectorLiteral)
    if err != nil {
        pgLogger.Errorf("Failed to upsert embedding for text %q: %v", text, err)
    } else {
        pgLogger.Debugf("Successfully upserted embedding for text %q", text)
    }
    return err
}

// Get retrieves the embedding for the given text.
// Returns sql.ErrNoRows if not found.
func (s *PostgresStore) Get(ctx context.Context, text string) ([]float32, error) {
    pgLogger.Debugf("Querying embedding for text %q from pgvector (dimension=%d)", text, s.dimension)
    row := s.db.QueryRowContext(ctx,
        `SELECT embedding FROM embeddings WHERE text = $1`, text)
    var vec string
    if err := row.Scan(&vec); err != nil {
        pgLogger.Infof("No embedding found for text %q in pgvector: %v", text, err)
        return nil, err
    }
    pgLogger.Debugf("Retrieved embedding for text %q from pgvector", text)
    parsed, err := parseVectorLiteral(vec)
    if err != nil {
        pgLogger.Errorf("Failed to parse vector literal for text %q: %v", text, err)
        return nil, err
    }
    return parsed, nil
}

// SearchByEmbedding returns the top k stored texts closest to the given embedding.
func (s *PostgresStore) SearchByEmbedding(ctx context.Context, embedding []float32, k int) ([]SimilarItem, error) {
    vectorLiteral := toVectorLiteral(embedding)
    query := fmt.Sprintf(
        `SELECT text, embedding <=> $1::vector AS distance
FROM embeddings
ORDER BY embedding <=> $1::vector
LIMIT %d`, k)
    rows, err := s.db.QueryContext(ctx, query, vectorLiteral)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var items []SimilarItem
    for rows.Next() {
        var it SimilarItem
        if err := rows.Scan(&it.Text, &it.Distance); err != nil {
            return nil, err
        }
        items = append(items, it)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return items, nil
}

// toVectorLiteral formats a []float32 as a pgvector literal "[x1,x2,...]".
func toVectorLiteral(vec []float32) string {
    parts := make([]string, len(vec))
    for i, v := range vec {
        parts[i] = fmt.Sprintf("%g", v)
    }
    return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// parseVectorLiteral parses a pgvector literal "[x1,x2,...]" into a []float32.
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
