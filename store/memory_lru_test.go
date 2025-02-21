package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLRUCache_Get tests the behavior of the Get method in the LRU cache for various scenarios, including cache hits and misses.
func TestLRUCache_Get(t *testing.T) {
	cache := NewLRUCache[string](2)

	tests := []struct {
		name          string
		existingKey   string
		existingValue string
		queryKey      string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "key cacheValid",
			existingKey:   "key1",
			existingValue: "value1",
			queryKey:      "key1",
			expectedValue: "value1",
			expectedFound: true,
		},
		{
			name:          "key does not exist",
			existingKey:   "key1",
			existingValue: "value1",
			queryKey:      "key2",
			expectedValue: "",
			expectedFound: false,
		},
	}

	err := cache.Set(context.Background(), "key1", "value1")
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found, err := cache.Get(context.Background(), tt.queryKey)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

// TestLRUCache_Get2 verifies the behavior of the LRU cache's Get method when attempting to retrieve an evicted key.
func TestLRUCache_Get2(t *testing.T) {
	cache := NewLRUCache[string](2)

	tests := []struct {
		name          string
		existingKey   string
		existingValue string
		queryKey      string
		expectedValue string
		expectedFound bool
	}{

		{
			name:          "key evicted",
			existingKey:   "key3",
			existingValue: "value3",
			queryKey:      "key1", // "key1" was evicted
			expectedValue: "",
			expectedFound: false,
		},
	}

	err := cache.Set(context.Background(), "key1", "value1")
	assert.NoError(t, err)

	err = cache.Set(context.Background(), "key2", "value2")
	assert.NoError(t, err)

	err = cache.Set(context.Background(), "key3", "value3") // Evicts "key1" because capacity is 2
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found, err := cache.Get(context.Background(), tt.queryKey)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

// TestLRUCache_Set tests the Set method of the LRU cache to verify key insertion, updates, and retrieval functionality.
func TestLRUCache_Set(t *testing.T) {
	cache := NewLRUCache[string](2)

	tests := []struct {
		name          string
		key           string
		value         string
		queryKey      string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "key added",
			key:           "key1",
			value:         "value1",
			queryKey:      "key1",
			expectedValue: "value1",
			expectedFound: true,
		},
		{
			name:          "key updated",
			key:           "key1",
			value:         "new_value",
			queryKey:      "key1",
			expectedValue: "new_value",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(context.Background(), tt.key, tt.value)
			assert.NoError(t, err)

			value, found, _ := cache.Get(context.Background(), tt.queryKey)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

// TestNewLRUCache tests the creation and behavior of the LRU cache, including eviction and retrieval of cached entries.
func TestNewLRUCache(t *testing.T) {
	t.Run("default size behavior", func(t *testing.T) {
		cache := NewLRUCache[string](2)
		err := cache.Set(context.Background(), "key1", "value1")
		assert.NoError(t, err)

		err = cache.Set(context.Background(), "key2", "value2")
		assert.NoError(t, err)
		err = cache.Set(context.Background(), "key3", "value3") // Evicts "key1"
		assert.NoError(t, err)

		_, found, _ := cache.Get(context.Background(), "key1")
		assert.False(t, found)

		value, found, _ := cache.Get(context.Background(), "key2")
		assert.True(t, found)
		assert.Equal(t, "value2", value)

		value, found, _ = cache.Get(context.Background(), "key3")
		assert.True(t, found)
		assert.Equal(t, "value3", value)
	})
}
