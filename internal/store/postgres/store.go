package postgres

import (
	"context"

	"github.com/jsojo/goshrt/internal/store"
)

// Store defines the PostgreSQL persistence contract for URL data.
// Implementations MUST use pgx/v5 and handle connection pooling internally.
type Store interface {
	// Create inserts a new URL record and returns it with populated fields (id, created_at, etc.).
	Create(ctx context.Context, url *store.URL) error

	// GetByShortCode retrieves a single active URL by its short code.
	// Returns nil, nil if not found.
	GetByShortCode(ctx context.Context, shortCode string) (*store.URL, error)

	// List returns a paginated list of URLs ordered by created_at DESC.
	List(ctx context.Context, offset, limit int) ([]*store.URL, error)

	// UpdateClicks sets the click count for a URL.
	UpdateClicks(ctx context.Context, shortCode string, clicks int64) error

	// SoftDelete marks a URL as inactive (is_active = false).
	SoftDelete(ctx context.Context, shortCode string) error

	// GetAllActive returns all URLs where is_active = true.
	// Used for cache seeding on startup.
	GetAllActive(ctx context.Context) ([]*store.URL, error)

	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error
}
