package echocache

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockCacher[T any] struct {
	cache   map[string]T
	mu      sync.Mutex
	getErr  error
	setErr  error
	keyFunc func(string) string
}

func (mc *mockCacher[T]) Get(ctx context.Context, key string) (value T, exists bool, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.getErr != nil {
		return value, false, mc.getErr
	}
	v, ok := mc.cache[key]
	return v, ok, nil
}

func (mc *mockCacher[T]) Set(ctx context.Context, key string, value T) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.setErr != nil {
		return mc.setErr
	}
	mc.cache[key] = value
	return nil
}

func (mc *mockCacher[T]) BuildKey(rawKey string) string {
	if mc.keyFunc != nil {
		return mc.keyFunc(rawKey)
	}
	return "key:" + rawKey
}

func TestEchoCache_Memoize(t *testing.T) {
	ctx := context.Background()

	t.Run("value_found_in_cache", func(t *testing.T) {
		mc := &mockCacher[string]{
			cache: map[string]string{
				"test": "cached value",
			},
		}
		cache := New[string](mc)
		value, exists, err := cache.Memoize(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed value", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "cached value", value)
	})

	t.Run("value_refreshed_and_cached", func(t *testing.T) {
		mc := &mockCacher[string]{cache: make(map[string]string)}
		cache := New[string](mc)
		value, exists, err := cache.Memoize(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed value", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "refreshed value", value)
		assert.Equal(t, "refreshed value", mc.cache["test"])
	})

	t.Run("value_refresh_fails", func(t *testing.T) {
		mc := &mockCacher[string]{cache: make(map[string]string)}
		cache := New[string](mc)
		value, exists, err := cache.Memoize(ctx, "test", func(ctx context.Context) (string, error) {
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
		cache := New[string](mc)
		value, exists, err := cache.Memoize(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed value", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "refreshed value", value)
		assert.Equal(t, "refreshed value", mc.cache["test"])
	})

	t.Run("cache_set_fails", func(t *testing.T) {
		mc := &mockCacher[string]{
			cache:  make(map[string]string),
			setErr: errors.New("cache set error"),
		}
		cache := New[string](mc)
		value, exists, err := cache.Memoize(ctx, "test", func(ctx context.Context) (string, error) {
			return "refreshed value", nil
		})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "refreshed value", value)
	})

}
