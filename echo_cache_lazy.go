package echocache

import (
	"context"
	"errors"
	"github.com/logocomune/echocache/store"
	"golang.org/x/sync/singleflight"
	"log/slog"
	"time"
)

// EchoCacheLazy is a lazy-refresh cache designed to handle stale-while-revalidate caching with generic type support.
// It queues refresh tasks to be processed later instead of blocking calls for cache updates.
// It leverages singleflight to prevent duplicate execution of refresh tasks for the same key.
// Refresh operations are managed with timeout and cancellation support for efficient processing.
// This type is suitable for scenarios where background cache updates improve application performance.
type EchoCacheLazy[T any] struct {
	store          store.StaleWhileRevalidateCache[T]
	sf             singleflight.Group
	queue          chan refreshTask[T]
	ctx            context.Context
	cancel         context.CancelFunc
	refreshTimeout time.Duration
}

// NewLazyEchoCache initializes a lazy echo cache with a specified stale-while-revalidate cacher and refresh timeout.
// It starts a background goroutine to handle refresh tasks and returns a pointer to the configured EchoCacheLazy instance.
func NewLazyEchoCache[T any](cacher store.StaleWhileRevalidateCache[T], refreshTimeout time.Duration) *EchoCacheLazy[T] {
	ctx, cancel := context.WithCancel(context.Background())

	lazyCache := EchoCacheLazy[T]{
		store:          cacher,
		sf:             singleflight.Group{},
		queue:          make(chan refreshTask[T], 1000),
		ctx:            ctx,
		cancel:         cancel,
		refreshTimeout: refreshTimeout,
	}
	go func() {

		for {
			select {
			case task := <-lazyCache.queue:
				_, _, _ = lazyCache.processRefreshTask(task, refreshTimeout)
			case <-lazyCache.ctx.Done():
				return

			}
		}

	}()

	return &lazyCache
}

// ShutdownLazyRefresh gracefully shuts down the refresh process by canceling the context and closing the task queue.
func (ec *EchoCacheLazy[T]) ShutdownLazyRefresh() {
	ec.cancel()
	close(ec.queue)
}

// FetchWithLazyRefresh retrieves a cached value or computes a new value if missing, scheduling a lazy refresh if needed.
// It uses a key to fetch a value from the cache and utilizes a provided function to refresh the value when necessary.
// If the cached value exists but is older than the lazy refresh interval, a refresh task is sent to the queue.
// If the value is missing or an error occurs during retrieval, a new value is computed immediately.
// Returns the cached or computed value, a boolean indicating cache hit, and an error if any.
func (ec *EchoCacheLazy[T]) FetchWithLazyRefresh(ctx context.Context, key string, refreshFn store.RefreshFunc[T], lazyRefreshInterval time.Duration) (T, bool, error) {

	// Attempt to retrieve the resultValue from the cache.
	value, exists, err := ec.store.Get(ctx, key)

	now := time.Now()
	if exists {
		if value.CreatedAt.Add(lazyRefreshInterval).Before(now) {
			slog.Info("Send task to queue")
			select {

			case ec.queue <- refreshTask[T]{
				key:         key,
				computeFunc: refreshFn,
				requestId:   randString(10),
			}:
			default:
				slog.Warn("processRefreshTask: queue is full, task dropped", slog.String("key", key))
			}
		}
		return value.Value, true, nil
	}
	if err != nil {
		// Log the error but proceed with computation.
		slog.Warn("Cannot get resultValue from cache", slog.String("error", err.Error()), slog.String("cacheKey", key))
	}

	task := refreshTask[T]{
		key:         key,
		computeFunc: refreshFn,
		requestId:   randString(10),
	}
	return ec.processRefreshTask(task, lazyRefreshInterval)

}

// processRefreshTask handles the computation and caching of a value, respecting the provided refresh timeout settings.
// It uses singleflight to ensure only one computation per key is performed and updates the cache if successful.
func (ec *EchoCacheLazy[T]) processRefreshTask(task refreshTask[T], refreshTimeout time.Duration) (T, bool, error) {
	var zeroValue T

	taskContext, cancel := context.WithTimeout(ec.ctx, ec.refreshTimeout)
	defer cancel()
	sfResult, sfErr, _ := ec.sf.Do(task.key, func() (interface{}, error) {
		res, err := task.computeFunc(taskContext)
		return singleFlightResult[T]{
			resultValue: res,
			createdAt:   time.Now(),
			requestId:   task.requestId,
		}, err
	})

	if sfErr != nil {
		slog.Error("processRefreshTask: failed to refresh resultValue", slog.String("key", task.key), slog.String("error", sfErr.Error()))
		return zeroValue, false, sfErr
	}

	// Validate the computed sfResult's type.
	resolvedValue, ok := sfResult.(singleFlightResult[T])
	if !ok {
		slog.Error("processRefreshTask: type assertion to singleFlightResult failed", slog.String("key", task.key))
		return zeroValue, false, errors.New("type assertion failed for computed resultValue")
	}
	if task.requestId == resolvedValue.requestId {

		cachedItem := store.StaleValue[T]{
			Value:     resolvedValue.resultValue,
			CreatedAt: resolvedValue.createdAt,
		}
		if err := ec.store.Set(taskContext, task.key, cachedItem); err != nil {
			// Log the error but still return the computed resultValue.
			slog.Warn("Failed to store resultValue in cache", slog.String("key", task.key), slog.String("error", err.Error()))
		}
	}
	return resolvedValue.resultValue, true, nil

}
