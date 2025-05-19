-- Initialize pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;
-- Ensure legacy embeddings table is recreated with correct dimensions
DROP TABLE IF EXISTS embeddings;

-- Create a table for storing embeddings (legacy format for backward compatibility)
CREATE TABLE IF NOT EXISTS embeddings (
    text TEXT PRIMARY KEY,
    embedding vector(%[1]d)
);

-- Create an index for the legacy embeddings table
CREATE INDEX IF NOT EXISTS idx_embeddings_embedding
    ON embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Create a table for storing prompt strategies with embeddings
CREATE TABLE IF NOT EXISTS prompt_strategies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    path VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    embedding vector(%[1]d),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create an index for vector similarity search on prompt strategies
CREATE INDEX IF NOT EXISTS prompt_strategies_embedding_idx 
    ON prompt_strategies USING ivfflat (embedding vector_cosine_ops) 
    WITH (lists = 100);

-- Create a table for caching query embeddings
CREATE TABLE IF NOT EXISTS embedding_cache (
    id SERIAL PRIMARY KEY,
    text_content TEXT NOT NULL UNIQUE,
    embedding vector(%[1]d),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create a table for logging classification requests
CREATE TABLE IF NOT EXISTS classification_logs (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    matched_strategy VARCHAR(255) NOT NULL,
    similarity_score FLOAT NOT NULL,
    classification_method VARCHAR(50) NOT NULL,
    processing_time_ms INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create a view for performance analytics
CREATE OR REPLACE VIEW classification_performance AS
SELECT 
    classification_method,
    COUNT(*) as request_count,
    AVG(similarity_score) as avg_similarity,
    AVG(processing_time_ms) as avg_processing_time,
    MIN(processing_time_ms) as min_processing_time,
    MAX(processing_time_ms) as max_processing_time,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY processing_time_ms) as p95_processing_time,
    DATE_TRUNC('hour', created_at) as hour
FROM classification_logs
GROUP BY classification_method, DATE_TRUNC('hour', created_at)
ORDER BY hour DESC;

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to update the updated_at timestamp
DROP TRIGGER IF EXISTS update_prompt_strategies_updated_at ON prompt_strategies;
CREATE TRIGGER update_prompt_strategies_updated_at
BEFORE UPDATE ON prompt_strategies
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create a function for cosine similarity search
CREATE OR REPLACE FUNCTION find_similar_strategies(query_embedding vector, similarity_threshold float, max_results integer)
RETURNS TABLE (
    id integer,
    name varchar,
    path varchar,
    content text,
    similarity float
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ps.id,
        ps.name,
        ps.path,
        ps.content,
        1 - (ps.embedding <=> query_embedding) as similarity
    FROM 
        prompt_strategies ps
    WHERE
        1 - (ps.embedding <=> query_embedding) > similarity_threshold
    ORDER BY 
        ps.embedding <=> query_embedding
    LIMIT max_results;
END;
$$ LANGUAGE plpgsql;

-- Create a migration function to move data from legacy table to new table
CREATE OR REPLACE FUNCTION migrate_legacy_embeddings()
RETURNS INTEGER AS $$
DECLARE
    count_migrated INTEGER;
BEGIN
    INSERT INTO embedding_cache (text_content, embedding)
    SELECT text, embedding
    FROM embeddings
    ON CONFLICT (text_content) DO NOTHING;
    
    GET DIAGNOSTICS count_migrated = ROW_COUNT;
    RETURN count_migrated;
END;
$$ LANGUAGE plpgsql;