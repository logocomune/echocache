package store

import (
	"context"
	"sync"
	"time"
)

// singleEntryCache represents a thread-safe cache for storing a single generic entry with a time-to-live (TTL) mechanism.
type singleEntryCache[T any] struct {
	cache       T
	ttl         time.Duration
	lastUpdated time.Time
	cacheValid  bool
	sync.RWMutex
}

// newSingleEntryCache initializes a single-entry cache with the specified time-to-live duration.
func newSingleEntryCache[T any](ttl time.Duration) *singleEntryCache[T] {
	return &singleEntryCache[T]{ttl: ttl}
}

// NewSingleCache creates a single-entry cache with the specified TTL, returning a generic Cacher interface instance.
func NewSingleCache[T any](ttl time.Duration) Cacher[T] {
	return newSingleEntryCache[T](ttl)
}

// NewStaleWhileRevalidateSingleCache creates a single-entry cache with a stale-while-revalidate pattern and a specified TTL.
func NewStaleWhileRevalidateSingleCache[T any](ttl time.Duration) StaleWhileRevalidateCache[T] {
	return newSingleEntryCache[StaleValue[T]](ttl)
}

// Get retrieves the cached value, a boolean indicating if the value exists, and an error if applicable.
// Returns an empty value if the cache is invalid or expired.
func (s *singleEntryCache[T]) Get(_ context.Context, _ string) (T, bool, error) {

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

// Set updates the cached value, marks it as valid, and sets the last updated timestamp.
func (s *singleEntryCache[T]) Set(_ context.Context, _ string, value T) error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	s.cache = value
	s.lastUpdated = time.Now()
	s.cacheValid = true
	return nil
}

// TryAcquireRefreshLock attempts to acquire a lock for refreshing the cache entry and returns true if successful.
func (l *singleEntryCache[T]) TryAcquireRefreshLock(_ context.Context, _ string, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

// ReleaseRefreshLock releases a previously acquired refresh lock, allowing other processes to proceed. Returns an error if unsuccessful.
func (l *singleEntryCache[T]) ReleaseRefreshLock(_ context.Context, _ string, _ string) error {
	return nil
}
