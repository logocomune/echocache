package echocache

import (
	"context"
	"sync"
	"time"
)

// _ ensures that SingleEntryCache satisfies the Cacher interface for any type T at compile time.
var _ Cacher[any] = &SingleEntryCache[any]{}

// SingleEntryCache is a generic cache supporting a single entry with a time-to-live (TTL) for expiration.
// It uses RWMutex for thread-safe operations, ensuring concurrent access safety.
type SingleEntryCache[T any] struct {
	cache       T
	ttl         time.Duration
	lastUpdated time.Time
	cacheValid  bool
	sync.RWMutex
}

// NewSingleCache creates and returns a new instance of SingleEntryCache with the specified time-to-live (TTL) duration.
func NewSingleCache[T any](ttl time.Duration) *SingleEntryCache[T] {

	return &SingleEntryCache[T]{
		ttl: ttl,
	}
}

// Get retrieves the cached value, its validity, and an error if applicable. Returns default value if cache is invalid or expired.
func (s *SingleEntryCache[T]) Get(_ context.Context, _ string) (T, bool, error) {

	var emptyValue T
	s.RWMutex.RLock()
	updated := s.lastUpdated
	ttl := s.ttl
	exists := s.cacheValid
	s.RWMutex.RUnlock()
	cached := s.cache
	if !exists {
		return emptyValue, false, nil
	}

	if time.Since(updated) > ttl {
		s.RWMutex.Lock()
		s.cacheValid = false
		s.cache = emptyValue
		s.RWMutex.Unlock()
		return emptyValue, false, nil
	}
	return cached, exists, nil

}

// Set updates the cache with a new value, marks it as valid, and records the current timestamp as the last updated time.
func (s *SingleEntryCache[T]) Set(_ context.Context, _ string, value T) error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	s.cache = value
	s.lastUpdated = time.Now()
	s.cacheValid = true
	return nil
}
