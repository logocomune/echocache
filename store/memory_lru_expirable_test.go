package store

import (
	"context"
	"testing"
	"time"
)

// TestNewLRUExpirableCache verifies the creation of a new LRU expirable cache and ensures it initializes correctly.
func TestNewLRUExpirableCache(t *testing.T) {
	cache := NewLRUExpirableCache[string](10, time.Minute)
	if cache == nil {
		t.Fatalf("expected non-nil cache")
	}
}

// TestNewStaleWhileRevalidateExpiringLRUCache validates the creation of a non-nil, correctly initialized cache instance.
func TestNewStaleWhileRevalidateExpiringLRUCache(t *testing.T) {
	cache := NewStaleWhileRevalidateExpiringLRUCache[string](10, time.Minute)
	if cache == nil {
		t.Fatalf("expected non-nil cache")
	}
}

// TestLRUExpirableCache_Get verifies the Get method functionality of an LRU expirable cache for existing and missing keys.
func TestLRUExpirableCache_Get(t *testing.T) {
	cache := newLRUExpirableCache[string](10, time.Minute)

	cache.cache.Add("key1", "value1")
	cache.cache.Add("key2", "value2")

	tests := []struct {
		name     string
		key      string
		expected string
		exists   bool
	}{
		{"key exists", "key1", "value1", true},
		{"key does not exist", "key3", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists, err := cache.Get(context.Background(), tt.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.exists || value != tt.expected {
				t.Fatalf("unexpected result, got value=%s, exists=%v; want value=%s, exists=%v", value, exists, tt.expected, tt.exists)
			}
		})
	}
}

// TestLRUExpirableCache_Set tests the Set method of the LRU expirable cache for correct adding and updating of key-value pairs.
func TestLRUExpirableCache_Set(t *testing.T) {
	cache := newLRUExpirableCache[string](10, time.Minute)

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"add key1", "key1", "value1"},
		{"add key2", "key2", "value2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(context.Background(), tt.key, tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			value, exists := cache.cache.Get(tt.key)
			if !exists || value != tt.value {
				t.Fatalf("unexpected value, got %s, want %s", value, tt.value)
			}
		})
	}
}

// TestLRUExpirableCache_TryAcquireRefreshLock tests the TryAcquireRefreshLock method to ensure locks can be acquired for specified keys.
func TestLRUExpirableCache_TryAcquireRefreshLock(t *testing.T) {
	cache := newLRUExpirableCache[string](10, time.Minute)

	tests := []struct {
		name string
		key  string
	}{
		{"acquire lock for key1", "key1"},
		{"acquire lock for key2", "key2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locked, err := cache.TryAcquireRefreshLock(context.Background(), tt.key, "randValue", time.Second)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !locked {
				t.Fatalf("expected to acquire lock, but failed")
			}
		})
	}
}

// TestLRUExpirableCache_ReleaseRefreshLock tests the ReleaseRefreshLock method to ensure proper release of refresh locks.
func TestLRUExpirableCache_ReleaseRefreshLock(t *testing.T) {
	cache := newLRUExpirableCache[string](10, time.Minute)

	tests := []struct {
		name      string
		key       string
		expectErr bool
	}{
		{"release lock for key1", "key1", false},
		{"release lock for key2", "key2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.ReleaseRefreshLock(context.Background(), tt.key, "randValue")
			if (err != nil) != tt.expectErr {
				t.Fatalf("unexpected error state, got err=%v, expectErr=%v", err, tt.expectErr)
			}
		})
	}
}
