package echocache

import (
	"context"
	lru "github.com/hashicorp/golang-lru/v2"
)

var _ Cacher[any] = &LRUCache[any]{}

// LRUCache is a generic cache implementation based on a least-recently-used (LRU) eviction policy.
type LRUCache[T any] struct {
	cache *lru.Cache[string, T]
}

// NewLRUCache creates and initializes a new LRUCache with the specified size and time-to-live (ttl) duration.
func NewLRUCache[T any](size int) *LRUCache[T] {
	c, _ := lru.New[string, T](size)

	return &LRUCache[T]{
		cache: c,
	}
}

// Get fetches the value associated with the specified key from the cache.
// Returns the value, a boolean indicating existence, and an error if any.
func (l *LRUCache[T]) Get(_ context.Context, key string) (value T, exists bool, err error) {
	value, exists = l.cache.Get(key)
	return value, exists, nil
}

// Set inserts a key-value pair into the LRUCache, updating an existing key if it already cacheValid.
func (l *LRUCache[T]) Set(_ context.Context, key string, value T) error {
	l.cache.Add(key, value)
	return nil
}
