-- migrations for pgvector and embeddings table

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS embeddings (
    text TEXT PRIMARY KEY,
    embedding vector(%d)
);

CREATE INDEX IF NOT EXISTS idx_embeddings_embedding
    ON embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
