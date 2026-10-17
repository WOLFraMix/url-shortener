package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(dbUrl string) (*Storage, error) {
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Save(originalURL, shortCode string) error {
	_, err := s.pool.Exec(context.Background(),
		"INSERT INTO links (original_url, short_code) VALUES ($1, $2)",
		originalURL, shortCode)
	return err
}

func (s *Storage) GetByShortCode(code string) (string, error) {
	var originalURL string
	err := s.pool.QueryRow(context.Background(),
		"SELECT original_url FROM links WHERE short_code = $1",
		code).Scan(&originalURL)
	return originalURL, err
}

func (s *Storage) GetByURL(originalURL string) (string, error) {
	var code string
	err := s.pool.QueryRow(context.Background(),
		"SELECT short_code FROM links WHERE original_url = $1",
		originalURL).Scan(&code)
	return code, err
}
