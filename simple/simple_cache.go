package simple

import (
	"context"
	"sync"
)

// SimpleCache is a generic thread-safe cache for storing a single value of any type.
// It uses read-write locking to manage concurrent access.
// The cache allows setting and retrieving a value with optional existence checking.
type SimpleCache[T any] struct {
	value *T
	sync.RWMutex
}

// NewSimpleCache creates and returns a new instance of SimpleCache with no initial value.
func NewSimpleCache[T any]() *SimpleCache[T] {
	return &SimpleCache[T]{}
}

// Get retrieves the cached value and indicates whether it exists without modifying the cache state.
// It locks the cache for reading and ensures thread-safe access.
// Returns the cached value, a boolean indicating existence, and an error if any occurs.
func (c *SimpleCache[T]) Get(_ context.Context, _ string) (value T, exists bool, err error) {
	c.RLock()
	defer c.RUnlock()

	if c.value == nil {
		return value, false, nil
	}

	return *c.value, true, nil
}

// Set sets the provided value into the SimpleCache. The operation is thread-safe with locking.
func (c *SimpleCache[T]) Set(_ context.Context, _ string, value T) error {
	c.Lock()
	defer c.Unlock()

	c.value = &value
	return nil
}
