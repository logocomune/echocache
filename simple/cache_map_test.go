package simple

import (
	"context"
	"testing"
)

func TestCacheMap_Get(t *testing.T) {
	type testCase[T any] struct {
		name           string
		setup          func(c *CacheMap[T])
		key            string
		expectedVal    T
		expectedExists bool
	}

	tests := []testCase[string]{
		{
			name: "key exists in map",
			setup: func(c *CacheMap[string]) {
				c.Set(context.Background(), "existingKey", "existingValue")
			},
			key:            "existingKey",
			expectedVal:    "existingValue",
			expectedExists: true,
		},
		{
			name:           "key does not exist in map",
			setup:          func(c *CacheMap[string]) {},
			key:            "nonExistentKey",
			expectedVal:    "",
			expectedExists: false,
		},
		{
			name: "empty map initialization",
			setup: func(c *CacheMap[string]) {
				// no setup, testing empty map
			},
			key:            "anyKey",
			expectedVal:    "",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewSimpleMap[string]()
			tt.setup(cache)

			val, exists, err := cache.Get(context.Background(), tt.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expectedVal {
				t.Errorf("expected value %q, got %q", tt.expectedVal, val)
			}
			if exists != tt.expectedExists {
				t.Errorf("expected existence %v, got %v", tt.expectedExists, exists)
			}
		})
	}
}

func TestCacheMap_Set(t *testing.T) {
	type testCase[T any] struct {
		name        string
		setup       func(c *CacheMap[T])
		key         string
		value       T
		expectedVal T
	}

	tests := []testCase[string]{
		{
			name:        "set new key",
			setup:       func(c *CacheMap[string]) {},
			key:         "newKey",
			value:       "newValue",
			expectedVal: "newValue",
		},
		{
			name: "update existing key",
			setup: func(c *CacheMap[string]) {
				c.Set(context.Background(), "existingKey", "oldValue")
			},
			key:         "existingKey",
			value:       "updatedValue",
			expectedVal: "updatedValue",
		},
		{
			name: "initialize nil map",
			setup: func(c *CacheMap[string]) {
				c.data = nil // explicitly nil to test initialization
			},
			key:         "newKey",
			value:       "newValue",
			expectedVal: "newValue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewSimpleMap[string]()
			tt.setup(cache)

			err := cache.Set(context.Background(), tt.key, tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			val, exists, err := cache.Get(context.Background(), tt.key)
			if err != nil {
				t.Fatalf("unexpected error on get: %v", err)
			}
			if !exists {
				t.Fatalf("expected key %q to exist, but it does not", tt.key)
			}
			if val != tt.expectedVal {
				t.Errorf("expected value %q, got %q", tt.expectedVal, val)
			}
		})
	}
}
