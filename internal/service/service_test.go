package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jsojo/goshrt/internal/store"
)

// mockStore implements Store interface
type mockStore struct {
	urls map[string]*store.URL
}

func newMockStore() *mockStore {
	return &mockStore{urls: make(map[string]*store.URL)}
}

func (m *mockStore) Create(_ context.Context, url *store.URL) error {
	if _, exists := m.urls[url.ShortCode]; exists {
		return errors.New("duplicate key")
	}
	m.urls[url.ShortCode] = url
	return nil
}

func (m *mockStore) GetByShortCode(_ context.Context, shortCode string) (*store.URL, error) {
	return m.urls[shortCode], nil
}

func (m *mockStore) List(_ context.Context, offset, limit int) ([]*store.URL, error) {
	var result []*store.URL
	for _, u := range m.urls {
		result = append(result, u)
	}
	if offset >= len(result) {
		return []*store.URL{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockStore) UpdateClicks(_ context.Context, shortCode string, clicks int64) error {
	if u, ok := m.urls[shortCode]; ok {
		u.Clicks += clicks
		return nil
	}
	return errors.New("not found")
}

func (m *mockStore) SoftDelete(_ context.Context, shortCode string) error {
	if u, ok := m.urls[shortCode]; ok {
		u.IsActive = false
		return nil
	}
	return errors.New("not found")
}

func (m *mockStore) GetAllActive(_ context.Context) ([]*store.URL, error) {
	var result []*store.URL
	for _, u := range m.urls {
		result = append(result, u)
	}
	return result, nil
}

func (m *mockStore) Ping(_ context.Context) error {
	return nil
}

// mockCache implements Cache interface
type mockCache struct {
	data   map[string]*store.URL
	clicks map[string]int64
}

func newMockCache() *mockCache {
	return &mockCache{
		data:   make(map[string]*store.URL),
		clicks: make(map[string]int64),
	}
}

func (m *mockCache) Get(_ context.Context, shortCode string) (*store.URL, error) {
	return m.data[shortCode], nil
}

func (m *mockCache) Set(_ context.Context, url *store.URL) error {
	m.data[url.ShortCode] = url
	return nil
}

func (m *mockCache) Delete(_ context.Context, shortCode string) error {
	delete(m.data, shortCode)
	return nil
}

func (m *mockCache) IncrementClicks(_ context.Context, shortCode string) (int64, error) {
	m.clicks[shortCode]++
	return m.clicks[shortCode], nil
}

func (m *mockCache) Ping(_ context.Context) error {
	return nil
}

// TESTS

func TestCreateURL(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	url, err := svc.CreateURL(context.Background(), "https://example.com", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if url.ShortCode == "" {
		t.Fatal("expected non-empty short code")
	}
	if url.OriginalURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", url.OriginalURL)
	}
	if !url.IsActive {
		t.Error("expected active URL")
	}
}

func TestCreateURL_WithCustomAlias(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	url, err := svc.CreateURL(context.Background(), "https://example.com", "mylink", 0)
	if err != nil {
		t.Fatal(err)
	}
	if url.ShortCode != "mylink" {
		t.Errorf("expected mylink, got %s", url.ShortCode)
	}
}

func TestCreateURL_WithTTL(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	url, err := svc.CreateURL(context.Background(), "https://example.com", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if url.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}
	if url.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestCreateURL_DuplicateAlias(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	_, err := svc.CreateURL(context.Background(), "https://example.com", "dup", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateURL(context.Background(), "https://other.com", "dup", 0)
	if err == nil {
		t.Fatal("expected error for duplicate alias")
	}
}

func TestGetURL_CacheHit(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	created, _ := svc.CreateURL(context.Background(), "https://example.com", "", 0)

	got, err := svc.GetURL(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected URL, got nil")
	}
	if got.OriginalURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", got.OriginalURL)
	}
}

func TestGetURL_NotFound(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	got, err := svc.GetURL(context.Background(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil for non-existent URL")
	}
}

func TestDeleteURL(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	created, _ := svc.CreateURL(context.Background(), "https://example.com", "", 0)

	err := svc.DeleteURL(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatal(err)
	}

	// Should be gone from cache
	got, _ := svc.cache.Get(context.Background(), created.ShortCode)
	if got != nil {
		t.Error("expected URL to be removed from cache")
	}
}

func TestListURLs(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	svc.CreateURL(context.Background(), "https://a.com", "", 0)
	svc.CreateURL(context.Background(), "https://b.com", "", 0)

	urls, err := svc.ListURLs(context.Background(), 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
}

func TestListURLs_Pagination(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	for i := 0; i < 5; i++ {
		svc.CreateURL(context.Background(), "https://example.com", "", 0)
	}

	page1, _ := svc.ListURLs(context.Background(), 1, 2)
	if len(page1) != 2 {
		t.Errorf("expected 2 on page 1, got %d", len(page1))
	}

	page3, _ := svc.ListURLs(context.Background(), 3, 2)
	if len(page3) != 1 {
		t.Errorf("expected 1 on page 3, got %d", len(page3))
	}
}

func TestIncrementClick(t *testing.T) {
	svc := New(newMockStore(), newMockCache())
	created, _ := svc.CreateURL(context.Background(), "https://example.com", "", 0)

	clicks, err := svc.IncrementClick(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatal(err)
	}
	if clicks != 1 {
		t.Errorf("expected 1 click, got %d", clicks)
	}

	clicks, _ = svc.IncrementClick(context.Background(), created.ShortCode)
	if clicks != 2 {
		t.Errorf("expected 2 clicks, got %d", clicks)
	}
}

func TestSyncClicks(t *testing.T) {
	mockSt := newMockStore()
	svc := New(mockSt, newMockCache())
	created, _ := svc.CreateURL(context.Background(), "https://example.com", "", 0)

	svc.IncrementClick(context.Background(), created.ShortCode)
	svc.IncrementClick(context.Background(), created.ShortCode)
	svc.IncrementClick(context.Background(), created.ShortCode)

	err := svc.SyncClicks(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Store should have updated clicks
	stored, _ := mockSt.GetByShortCode(context.Background(), created.ShortCode)
	if stored.Clicks != 3 {
		t.Errorf("expected 3 clicks synced, got %d", stored.Clicks)
	}
}

func TestSeedCache(t *testing.T) {
	mockSt := newMockStore()
	svc := New(mockSt, newMockCache())
	svc.CreateURL(context.Background(), "https://a.com", "", 0)
	svc.CreateURL(context.Background(), "https://b.com", "", 0)

	// Clear the cache
	svc.cache = newMockCache()

	count, err := svc.SeedCache(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 seeded, got %d", count)
	}
}
