package echocache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	cache.Set(context.Background(), "key1", "value1")
	//cache.Set(context.Background(), "key2", "value2")
	//	cache.Set(context.Background(), "key3", "value3") // Evicts "key1" because capacity is 2

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found, err := cache.Get(context.Background(), tt.queryKey)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

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

	cache.Set(context.Background(), "key1", "value1")
	cache.Set(context.Background(), "key2", "value2")
	cache.Set(context.Background(), "key3", "value3") // Evicts "key1" because capacity is 2

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found, err := cache.Get(context.Background(), tt.queryKey)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

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

func TestNewLRUCache(t *testing.T) {
	t.Run("default size behavior", func(t *testing.T) {
		cache := NewLRUCache[string](2)
		cache.Set(context.Background(), "key1", "value1")
		cache.Set(context.Background(), "key2", "value2")
		cache.Set(context.Background(), "key3", "value3") // Evicts "key1"

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
