package internal

// Config holds all CLI flag and argument values.
type Config struct {
	Dir       string  // Directory containing .md files
	Ext       string  // File extension
	Threshold float64 // Cosine similarity threshold

	DbDSN string // PostgreSQL DSN
	DbDim int    // Embedding vector dimension

	SeedOnly bool // Only seed strategies and exit
	UseDB    bool // Use database-based matching

	Query string // User query (positional arg)
}
