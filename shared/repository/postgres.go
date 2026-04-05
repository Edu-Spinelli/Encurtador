package repository

import (
	"database/sql"
	"time"

	"github.com/spinelli/encurtador-links/shared/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			short_code TEXT PRIMARY KEY,
			long_url   TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (r *PostgresRepository) Save(shortCode, longURL string) error {
	_, err := r.db.Exec(
		`INSERT INTO urls (short_code, long_url, created_at) VALUES ($1, $2, $3)`,
		shortCode, longURL, time.Now(),
	)
	return err
}

func (r *PostgresRepository) FindByShortCode(shortCode string) (*model.URL, error) {
	var u model.URL
	err := r.db.QueryRow(
		`SELECT short_code, long_url, created_at FROM urls WHERE short_code = $1`,
		shortCode,
	).Scan(&u.ShortCode, &u.LongURL, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
