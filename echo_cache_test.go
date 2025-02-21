package echocache

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockCacher is a generic struct that implements basic caching functionality for testing purposes.
// It includes methods to get, set, and build keys, and supports injecting errors for testing scenarios.
type mockCacher[T any] struct {
	cache   map[string]T
	mu      sync.Mutex
	getErr  error
	setErr  error
	keyFunc func(string) string
}

// Get retrieves a value associated with the given key from the cache, returning whether the key exists and any error encountered.
func (mc *mockCacher[T]) Get(ctx context.Context, key string) (value T, exists bool, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.getErr != nil {
		return value, false, mc.getErr
	}
	v, ok := mc.cache[key]
	return v, ok, nil
}

// Set stores the given value in the cache with the specified key and returns an error if the operation fails.
func (mc *mockCacher[T]) Set(ctx context.Context, key string, value T) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.setErr != nil {
		return mc.setErr
	}
	mc.cache[key] = value
	return nil
}

// BuildKey generates a cache key by applying a custom key function if defined, or prepends "key:" to the rawKey otherwise.
func (mc *mockCacher[T]) BuildKey(rawKey string) string {
	if mc.keyFunc != nil {
		return mc.keyFunc(rawKey)
	}
	return "key:" + rawKey
}

// TestEchoCache_Memoize tests the behavior of the EchoCache implementation with various caching scenarios and edge cases.
func TestEchoCache_Memoize(t *testing.T) {
	ctx := context.Background()

	t.Run("value_found_in_cache", func(t *testing.T) {
		mc := &mockCacher[string]{
			cache: map[string]string{
				"test": "cached resultValue",
			},
		}
		cache := NewEchoCache[string](mc)
		value, exists, err := cache.FetchWithCache(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed resultValue", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "cached resultValue", value)
	})

	t.Run("value_refreshed_and_cached", func(t *testing.T) {
		mc := &mockCacher[string]{cache: make(map[string]string)}
		cache := NewEchoCache[string](mc)
		value, exists, err := cache.FetchWithCache(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed resultValue", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "refreshed resultValue", value)
		assert.Equal(t, "refreshed resultValue", mc.cache["test"])
	})

	t.Run("value_refresh_fails", func(t *testing.T) {
		mc := &mockCacher[string]{cache: make(map[string]string)}
		cache := NewEchoCache[string](mc)
		value, exists, err := cache.FetchWithCache(ctx, "test", func(ctx context.Context) (string, error) {
			return "", errors.New("refresh error")
		})

		assert.Error(t, err)
		assert.False(t, exists)
		assert.Equal(t, "", value)
	})

	t.Run("cache_get_fails", func(t *testing.T) {
		mc := &mockCacher[string]{
			cache:  make(map[string]string),
			getErr: errors.New("cache get error"),
		}
		cache := NewEchoCache[string](mc)
		value, exists, err := cache.FetchWithCache(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed resultValue", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "refreshed resultValue", value)
		assert.Equal(t, "refreshed resultValue", mc.cache["test"])
	})

	t.Run("cache_set_fails", func(t *testing.T) {
		mc := &mockCacher[string]{
			cache:  make(map[string]string),
			setErr: errors.New("cache set error"),
		}
		cache := NewEchoCache[string](mc)
		value, exists, err := cache.FetchWithCache(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed resultValue", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "refreshed resultValue", value)
	})

}
