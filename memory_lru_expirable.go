package echocache

import (
	"context"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"time"
)

var _ Cacher[any] = &LRUExpirableCache[any]{}

// LRUExpirableCache represents a least-recently-used cache with expirable entries.
// It holds a generic type T and manages key-value pairs with expiration functionality.
type LRUExpirableCache[T any] struct {
	cache *expirable.LRU[string, T]
}

// NewLRUExplorableCache initializes a new LRU cache with a specified size and time-to-live (TTL) for cached items.
func NewLRUExplorableCache[T any](size int, ttl time.Duration) *LRUExpirableCache[T] {
	return &LRUExpirableCache[T]{
		cache: expirable.NewLRU[string, T](size, nil, ttl),
	}
}

// Get retrieves the value associated with the specified key from the cache, indicating whether it cacheValid or not.
func (l *LRUExpirableCache[T]) Get(_ context.Context, key string) (value T, exists bool, err error) {
	value, exists = l.cache.Get(key)
	return value, exists, nil
}

// Set inserts a key-value pair into the LRU expirable cache. It overwrites the value if the key already cacheValid.
func (l *LRUExpirableCache[T]) Set(_ context.Context, key string, value T) error {
	l.cache.Add(key, value)
	return nil
}
