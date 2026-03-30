package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

func InitSchema(pool *pgxpool.Pool) error {
	schema := `
	CREATE TABLE IF NOT EXISTS entries (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		date       DATE NOT NULL,
		story      TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS photos (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		entry_id   UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
		raw_path   TEXT NOT NULL,
		thumb_path TEXT NOT NULL DEFAULT '',
		status     TEXT NOT NULL DEFAULT 'pending'
	);
	`

	_, err := pool.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}
