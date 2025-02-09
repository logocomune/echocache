package echocache

import (
	"context"
	"errors"
	"golang.org/x/sync/singleflight"
	"log/slog"
)

// RefreshFunc is a generic type representing a function that refreshes a value and returns it with a potential error.
type RefreshFunc[T any] func(ctx context.Context) (T, error)

// Cacher defines an interface for caching data, supporting generic types for flexibility.
type Cacher[T any] interface {
	Get(ctx context.Context, key string) (value T, exists bool, err error)
	Set(ctx context.Context, key string, value T) error
	BuildKey(key string) string
}

// EchoCache is a generic caching mechanism that wraps a Cacher implementation with singleflight and a TTL configuration.
type EchoCache[T any] struct {
	db Cacher[T]
	sf singleflight.Group
}

// New creates a new instance of EchoCache with the provided Cacher and time-to-live (TTL) settings.
func New[T any](r Cacher[T]) *EchoCache[T] {

	return &EchoCache[T]{
		db: r,
		sf: singleflight.Group{},
	}
}

// Memoize retrieves a value from the cache or calculates and stores it using the provided RefreshFunc if it doesn't exist.
// Returns the value, a boolean indicating if the value was retrieved from the cache, and an error if any occurred.
func (ec *EchoCache[T]) Memoize(ctx context.Context, key string, refreshFn RefreshFunc[T]) (T, bool, error) {
	var emptyValue T
	cacheKey := ec.db.BuildKey(key)

	// Attempt to retrieve the value from the cache.
	value, exists, err := ec.db.Get(ctx, cacheKey)
	if exists {
		return value, true, nil
	}
	if err != nil {
		// Log the error but proceed with computation.
		slog.Warn("Cannot get value from cache", slog.String("error", err.Error()), slog.String("cacheKey", cacheKey))
	}

	// Use singleflight to ensure only one computation is made per key.
	result, errGroup, _ := ec.sf.Do(cacheKey, func() (interface{}, error) {
		return refreshFn(ctx)
	})
	if errGroup != nil {
		return emptyValue, false, errGroup
	}

	// Validate the computed result's type.
	computedValue, ok := result.(T)

	if !ok {
		return emptyValue, false, errors.New("type assertion failed for computed value")
	}

	// Save the computed value in the cache.
	if err := ec.db.Set(ctx, cacheKey, computedValue); err != nil {
		// Log the error but still return the computed value.
		slog.Warn("Failed to store value in cache", slog.String("key", cacheKey), slog.String("error", err.Error()))
	}

	return computedValue, true, nil
}
