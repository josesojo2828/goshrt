package store

import "time"

// URL represents a shortened URL entry stored in PostgreSQL and cached in Redis.
type URL struct {
	ID           string     `json:"id" db:"id"`
	ShortCode    string     `json:"short_code" db:"short_code"`
	OriginalURL  string     `json:"original_url" db:"original_url"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	Clicks       int64      `json:"clicks" db:"clicks"`
	LastAccessed *time.Time `json:"last_accessed,omitempty" db:"last_accessed"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
}
