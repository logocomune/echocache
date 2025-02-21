package store

import (
	"context"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"time"
)

// lruExpirableCache is a generic cache structure with LRU eviction and expiration time support.
// It wraps an expirable LRU cache implementation with string keys and generic type values.
// Provides methods for getting, setting, and managing refresh locks on cached items.
type lruExpirableCache[T any] struct {
	cache *expirable.LRU[string, T]
}

// NewLRUExpirableCache creates a new LRU cache with a specified size and time-to-live (TTL) for each entry.
func NewLRUExpirableCache[T any](size int, ttl time.Duration) Cacher[T] {
	return newLRUExpirableCache[T](size, ttl)
}

// NewStaleWhileRevalidateExpiringLRUCache creates a new LRU-based cache with support for stale-while-revalidate and expirable items.
// The cache allows a maximum of `size` items and applies a time-to-live duration defined by `ttl` for stored data.
func NewStaleWhileRevalidateExpiringLRUCache[T any](size int, ttl time.Duration) StaleWhileRevalidateCache[T] {
	return newLRUExpirableCache[StaleValue[T]](size, ttl)
}

// newLRUExpirableCache creates a new expirable LRU cache with a specified size and time-to-live duration.
func newLRUExpirableCache[T any](size int, ttl time.Duration) *lruExpirableCache[T] {
	return &lruExpirableCache[T]{
		cache: expirable.NewLRU[string, T](size, nil, ttl),
	}
}

// Get retrieves the value associated with the given key from the cache. Returns the value, if it exists, and any error encountered.
func (l *lruExpirableCache[T]) Get(_ context.Context, key string) (value T, exists bool, err error) {
	value, exists = l.cache.Get(key)
	return value, exists, nil
}

// Set adds a key-value pair to the cache. If the key already exists, its value is updated. Returns an error if the operation fails.
func (l *lruExpirableCache[T]) Set(_ context.Context, key string, value T) error {
	l.cache.Add(key, value)
	return nil
}

// TryAcquireRefreshLock attempts to acquire a refresh lock for the specified key and duration.
// Returns true if the lock is successfully acquired, false otherwise.
// An error is returned if the lock acquisition fails unexpectedly.
func (l *lruExpirableCache[T]) TryAcquireRefreshLock(_ context.Context, _ string, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

// ReleaseRefreshLock releases a previously acquired refresh lock for a given key and returns an error if it fails.
func (l *lruExpirableCache[T]) ReleaseRefreshLock(_ context.Context, _ string, _ string) error {
	return nil
}
