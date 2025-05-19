package promptlog

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// PostgresLogger stores prompt and response logs in PostgreSQL.
type PostgresLogger struct {
	db *sql.DB
}

// NewPostgresLogger connects to PostgreSQL and ensures the logging table exists.
func NewPostgresLogger(dsn string) (Logger, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS llm_logs (
			id SERIAL PRIMARY KEY,
			prompt TEXT,
			response TEXT,
			token TEXT,
			timestamp TIMESTAMPTZ
		)
	`)
	if err != nil {
		return nil, err
	}
	return &PostgresLogger{db: db}, nil
}

// LogPrompt inserts a prompt entry without a response.
func (l *PostgresLogger) LogPrompt(p, t string, ts time.Time) error {
	_, err := l.db.Exec(
		`INSERT INTO llm_logs (prompt, token, timestamp) VALUES ($1, $2, $3)`,
		p, t, ts,
	)
	return err
}

// LogResponse inserts a full prompt + response entry.
func (l *PostgresLogger) LogResponse(p, r, t string, ts time.Time) error {
	_, err := l.db.Exec(
		`INSERT INTO llm_logs (prompt, response, token, timestamp) VALUES ($1, $2, $3, $4)`,
		p, r, t, ts,
	)
	return err
}
