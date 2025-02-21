package echocache

import (
	"context"
	"errors"
	"github.com/logocomune/echocache/store"
	"golang.org/x/sync/singleflight"
	"log/slog"
	"time"
)

// NeverExpire represents a duration of 100 years, effectively used to denote a value that should never expire.
const (
	NeverExpire = time.Hour * 24 * 365 * 100
)

// EchoCache is a generic caching mechanism that integrates singleflight to prevent redundant computations.
// EchoCache uses a Cacher interface for data storage and retrieval, supporting custom refresh functions for cache misses.
// EchoCache ensures only one computation per key occurs simultaneously to optimize concurrent operations.
type EchoCache[T any] struct {
	store store.Cacher[T]
	sf    singleflight.Group
}

// NewEchoCache creates a new EchoCache instance to enable caching with optional singleflight for concurrent requests.
func NewEchoCache[T any](cacher store.Cacher[T]) *EchoCache[T] {

	return &EchoCache[T]{
		store: cacher,
		sf:    singleflight.Group{},
	}
}

// FetchWithCache retrieves a cached value by key or computes it using a given refresh function, caching the result for future use.
// Returns the value, a boolean indicating if it was found or computed, and an error if computation or retrieval fails.
func (ec *EchoCache[T]) FetchWithCache(ctx context.Context, key string, refreshFn store.RefreshFunc[T]) (T, bool, error) {
	var zeroValue T

	// Attempt to retrieve the resultValue from the cache.
	value, exists, err := ec.store.Get(ctx, key)
	if exists {
		return value, true, nil
	}
	if err != nil {
		// Log the error but proceed with computation.
		slog.Warn("Cannot get resultValue from cache", slog.String("error", err.Error()), slog.String("cacheKey", key))
	}

	requestId := randString(10)
	// Use singleflight to ensure only one computation is made per key.
	sfResult, sfErr, _ := ec.sf.Do(key, func() (interface{}, error) {
		v, e := refreshFn(ctx)
		res := singleFlightResult[T]{
			resultValue: v,
			createdAt:   time.Now(),
			requestId:   requestId,
		}

		return res, e
	})
	if sfErr != nil {
		return zeroValue, false, sfErr
	}

	// Validate the computed sfResult's type.
	resolvedValue, ok := sfResult.(singleFlightResult[T])

	if !ok {
		return zeroValue, false, errors.New("type assertion failed for computed resultValue")
	}

	if resolvedValue.requestId == requestId {
		// Save the computed resultValue in the cache.
		if err := ec.store.Set(ctx, key, resolvedValue.resultValue); err != nil {
			// Log the error but still return the computed resultValue.
			slog.Warn("Failed to store resultValue in cache", slog.String("key", key), slog.String("error", err.Error()))
		}

	}
	return resolvedValue.resultValue, true, nil
}
