package store

import (
	"context"
	"time"
)

// RefreshFunc defines a function type for refreshing or computing a value in a cache, returning the value and an error.
type RefreshFunc[T any] func(ctx context.Context) (T, error)

// Cacher is a generic interface for caching, allowing storage and retrieval of values with methods to handle cache entries.
type Cacher[T any] interface {
	Get(ctx context.Context, key string) (value T, exists bool, err error)
	Set(ctx context.Context, key string, value T) error
}

// StaleWhileRevalidateCache is a generic interface for a cache implementing stale-while-revalidate pattern.
// The cache is capable of storing and retrieving stale values while allowing background refresh of data.
// It embeds Cacher for basic caching operations and provides methods for managing refresh locks.
type StaleWhileRevalidateCache[T any] interface {
	Cacher[StaleValue[T]]
	TryAcquireRefreshLock(ctx context.Context, key string, randValue string, ttl time.Duration) (bool, error)
	ReleaseRefreshLock(ctx context.Context, key string, randValue string) error
}

// StaleValue represents a value associated with a timestamp indicating when it was created.
type StaleValue[T any] struct {
	Value     T
	CreatedAt time.Time
}
