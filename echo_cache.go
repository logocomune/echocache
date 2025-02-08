package echocache

import (
	"context"
	"log/slog"
	"time"
)

// RefreshFunc defines a function type that takes a context and returns a value of type T and an error.
type RefreshFunc[T any] func(ctx context.Context) (T, error)

// EchoCache is a generic cache wrapper that supports automatic data refresh based on a specified TTL duration.
// It uses a refresh function to retrieve new data and a channel for managing asynchronous refresh requests.
// The cache ensures data is either fetched from the underlying cache or refreshed when a cache miss occurs.
type EchoCache[T any] struct {
	refreshF    RefreshFunc[T]
	ttl         time.Duration
	lastRefresh time.Time
	cache       Cache[T]
	refreshChan chan string
}

// Cache is a generic interface for caching operations with methods to get and set values associated with a key.
type Cache[T any] interface {
	Get(ctx context.Context, key string) (value T, exists bool, err error)
	Set(ctx context.Context, key string, value T) error
}

// New initializes and returns an instance of EchoCache with the provided Cache, RefreshFunc, and TTL for cache entries.
func New[T any](c Cache[T], rf RefreshFunc[T], ttl time.Duration) *EchoCache[T] {
	refreshChan := make(chan string)
	e := EchoCache[T]{
		refreshF:    rf,
		ttl:         ttl,
		cache:       c,
		refreshChan: refreshChan,
	}
	go func() {
		for key := range refreshChan {
			t, err := rf(context.Background())
			if err != nil {
				slog.Warn("Error refreshing cache.", slog.String("error", err.Error()))
				continue
			}
			err = c.Set(context.Background(), key, t)
			if err != nil {
				slog.Warn("Error saving in cache.", slog.String("error", err.Error()))
				continue
			}
			e.lastRefresh = time.Now()
		}

	}()

	return &e
}

// Get retrieves the value associated with the specified key from the cache or refreshes it if not found or expired.
// Returns the value, a boolean indicating existence, and an error if any issue occurs during retrieval or refresh.
func (c *EchoCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	value, exists, err := c.cache.Get(ctx, key)
	if err != nil {

		return value, false, err
	}
	if exists {
		return value, true, nil
	}
	value, err = c.refreshF(ctx)
	if err != nil {
		return value, false, err
	}
	err = c.cache.Set(ctx, key, value)
	if err != nil {
		slog.Warn("Error saving in cache.", slog.String("error", err.Error()))
	}
	if time.Since(c.lastRefresh) > c.ttl {
		c.Refresh(key)
	}

	return value, true, nil

}

// Refresh attempts to pass the provided key to the refresh channel, enabling cache updates in a non-blocking manner.
func (c *EchoCache[T]) Refresh(key string) {
	select {
	case c.refreshChan <- key:
	default:
	}
}
