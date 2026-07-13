package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/jsojo/goshrt/internal/store"
)

type pgStore struct {
	db *sqlx.DB
}

func NewStore(dsn string) (*pgStore, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &pgStore{db: db}, nil
}

func (s *pgStore) Create(ctx context.Context, url *store.URL) error {
	query := `INSERT INTO urls (short_code, original_url, created_at, updated_at, expires_at, is_active)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.ExecContext(ctx, query,
		url.ShortCode, url.OriginalURL, url.CreatedAt, url.UpdatedAt, url.ExpiresAt, url.IsActive)
	return err
}

func (s *pgStore) GetByShortCode(ctx context.Context, shortCode string) (*store.URL, error) {
	var url store.URL
	query := `SELECT id, short_code, original_url, created_at, updated_at, clicks, last_accessed, expires_at, is_active
			  FROM urls WHERE short_code = $1 AND is_active = true`
	err := s.db.GetContext(ctx, &url, query, shortCode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *pgStore) List(ctx context.Context, offset, limit int) ([]*store.URL, error) {
	var urls []*store.URL
	query := `SELECT id, short_code, original_url, created_at, updated_at, clicks, last_accessed, expires_at, is_active
			  FROM urls WHERE is_active = true
			  ORDER BY created_at DESC
			  OFFSET $1 LIMIT $2`
	err := s.db.SelectContext(ctx, &urls, query, offset, limit)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *pgStore) UpdateClicks(ctx context.Context, shortCode string, clicks int64) error {
	query := `UPDATE urls SET clicks = clicks + $1, last_accessed = NOW() WHERE short_code = $2 AND is_active = true`
	_, err := s.db.ExecContext(ctx, query, clicks, shortCode)
	return err
}

func (s *pgStore) SoftDelete(ctx context.Context, shortCode string) error {
	query := `UPDATE urls SET is_active = false, updated_at = NOW() WHERE short_code = $1`
	_, err := s.db.ExecContext(ctx, query, shortCode)
	return err
}

func (s *pgStore) GetAllActive(ctx context.Context) ([]*store.URL, error) {
	var urls []*store.URL
	query := `SELECT id, short_code, original_url, created_at, updated_at, clicks, last_accessed, expires_at, is_active
			  FROM urls
			  WHERE is_active = true AND (expires_at IS NULL OR expires_at > NOW())`
	err := s.db.SelectContext(ctx, &urls, query)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *pgStore) Close() error {
	return s.db.Close()
}

func (s *pgStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
