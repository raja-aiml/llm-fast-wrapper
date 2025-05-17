package logs

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type PostgresLogger struct{ db *sql.DB }

func NewPostgresLogger(dsn string) (Logger, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS llm_logs (id SERIAL PRIMARY KEY,prompt TEXT,token TEXT,timestamp TIMESTAMPTZ)`)
	if err != nil {
		return nil, err
	}
	return &PostgresLogger{db: db}, nil
}

func (l *PostgresLogger) LogPrompt(p, t string, ts time.Time) error {
	_, err := l.db.Exec(`INSERT INTO llm_logs (prompt, token, timestamp) VALUES ($1,$2,$3)`, p, t, ts)
	return err
}
