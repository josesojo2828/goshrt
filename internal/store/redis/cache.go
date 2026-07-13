package redis

import (
	"context"

	"github.com/jsojo/goshrt/internal/store"
)

// Cache defines the Redis caching contract for URL data.
// Implementations MUST use go-redis/v9 and handle connection pooling internally.
type Cache interface {
	// Get retrieves a URL from cache by short code.
	// Returns nil, nil if not found (cache miss).
	Get(ctx context.Context, shortCode string) (*store.URL, error)

	// Set stores a URL in cache with appropriate TTL.
	Set(ctx context.Context, url *store.URL) error

	// Delete removes a URL from cache by short code.
	Delete(ctx context.Context, shortCode string) error

	// IncrementClicks atomically increments the click counter for a URL.
	// Returns the new click count after increment.
	IncrementClicks(ctx context.Context, shortCode string) (int64, error)

	// Ping verifies the Redis connection is alive.
	Ping(ctx context.Context) error
}
