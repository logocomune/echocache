package echocache

import (
	"github.com/logocomune/echocache/store"
	"time"
)

// singleFlightResult represents the result of a singleflight operation, including the value, creation time, and request ID.
type singleFlightResult[T any] struct {
	resultValue T
	createdAt   time.Time
	requestId   string
}

// refreshTask represents a task for refreshing a cache entry using a specified compute function.
type refreshTask[T any] struct {
	key         string
	computeFunc store.RefreshFunc[T]
	requestId   string
}
