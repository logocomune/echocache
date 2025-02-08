package simple

import (
	"context"
	"sync"
)

// CacheMap is a thread-safe generic map for storing key-value pairs with read/write locking.
type CacheMap[T any] struct {
	data map[string]T
	sync.RWMutex
}

// NewSimpleMap creates and returns a new instance of CacheMap initialized with an empty map.
func NewSimpleMap[T any]() *CacheMap[T] {
	return &CacheMap[T]{
		data: make(map[string]T),
	}
}

// Get retrieves the value associated with the given key from the CacheMap, returning if it exists and any error occurred.
func (c *CacheMap[T]) Get(_ context.Context, key string) (value T, exists bool, err error) {
	c.RLock()
	defer c.RUnlock()
	value, exists = c.data[key]
	return value, exists, nil
}

// Set inserts or updates the value associated with the specified key in the CacheMap. Thread-safe for concurrent use.
func (c *CacheMap[T]) Set(_ context.Context, key string, value T) error {
	c.Lock()
	defer c.Unlock()
	if c.data == nil {
		c.data = make(map[string]T)
	}
	c.data[key] = value
	return nil
}
