package store

import (
	"context"
	lru "github.com/hashicorp/golang-lru/v2"
	"time"
)

// lruCache is a generic wrapper around an LRU cache for storing and retrieving key-value pairs in a thread-safe manner.
type lruCache[T any] struct {
	cache *lru.Cache[string, T]
}

// NewLRUCache creates a new instance of a generic LRU cache with the specified size and returns it as a Cacher interface.
func NewLRUCache[T any](size int) Cacher[T] {
	c, _ := lru.New[string, T](size)

	return &lruCache[T]{
		cache: c,
	}
}

// NewStaleWhileRevalidateLRUCache creates a new LRU-based StaleWhileRevalidateCache with a specified size.
func NewStaleWhileRevalidateLRUCache[T any](size int) StaleWhileRevalidateCache[T] {
	c, _ := lru.New[string, StaleValue[T]](size)

	return &lruCache[StaleValue[T]]{
		cache: c,
	}
}

// Get retrieves the value associated with the given key from the cache.
// It returns the value, a boolean indicating if the key exists, and an error which is always nil.
func (l *lruCache[T]) Get(_ context.Context, key string) (value T, exists bool, err error) {
	value, exists = l.cache.Get(key)
	return value, exists, nil
}

// Set inserts a key-value pair into the LRU cache, potentially evicting an older entry, and returns an error if any occurs.
func (l *lruCache[T]) Set(_ context.Context, key string, value T) error {
	l.cache.Add(key, value)
	return nil
}

// TryAcquireRefreshLock attempts to acquire a refresh lock for the specified key, returning true if successful.
func (l *lruCache[T]) TryAcquireRefreshLock(_ context.Context, _ string, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

// ReleaseRefreshLock releases a previously acquired refresh lock for a cache key if applicable. Always returns nil.
func (l *lruCache[T]) ReleaseRefreshLock(_ context.Context, _ string, _ string) error {
	return nil
}
