package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/jsojo/goshrt/internal/store"
)

const redisKeyPrefix = "url:"

type redisCache struct {
	client *goredis.Client
}

func NewCache(addr, password string) (*redisCache, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	return &redisCache{client: client}, nil
}

func (c *redisCache) redisKey(shortCode string) string {
	return redisKeyPrefix + shortCode
}

func (c *redisCache) Get(ctx context.Context, shortCode string) (*store.URL, error) {
	fields, err := c.client.HGetAll(ctx, c.redisKey(shortCode)).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, nil
	}

	clicks, _ := strconv.ParseInt(fields["clicks"], 10, 64)

	url := &store.URL{
		ShortCode:   shortCode,
		OriginalURL: fields["original_url"],
		Clicks:      clicks,
	}

	if t, ok := fields["created_at"]; ok && t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			url.CreatedAt = parsed
		}
	}
	if t, ok := fields["last_accessed"]; ok && t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			url.LastAccessed = &parsed
		}
	}
	if t, ok := fields["expires_at"]; ok && t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			url.ExpiresAt = &parsed
		}
	}

	return url, nil
}

func (c *redisCache) Set(ctx context.Context, url *store.URL) error {
	key := c.redisKey(url.ShortCode)
	fields := map[string]string{
		"original_url": url.OriginalURL,
		"created_at":   url.CreatedAt.Format(time.RFC3339),
		"clicks":       strconv.FormatInt(url.Clicks, 10),
	}
	if url.LastAccessed != nil {
		fields["last_accessed"] = url.LastAccessed.Format(time.RFC3339)
	}
	if url.ExpiresAt != nil {
		fields["expires_at"] = url.ExpiresAt.Format(time.RFC3339)
	}

	if err := c.client.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("hset cache: %w", err)
	}

	if url.ExpiresAt != nil {
		ttl := time.Until(*url.ExpiresAt)
		if ttl > 0 {
			if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
				return fmt.Errorf("expire cache: %w", err)
			}
		}
	}

	return nil
}

func (c *redisCache) Delete(ctx context.Context, shortCode string) error {
	return c.client.Del(ctx, c.redisKey(shortCode)).Err()
}

func (c *redisCache) IncrementClicks(ctx context.Context, shortCode string) (int64, error) {
	return c.client.HIncrBy(ctx, c.redisKey(shortCode), "clicks", 1).Result()
}

func (c *redisCache) Close() error {
	return c.client.Close()
}

func (c *redisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
