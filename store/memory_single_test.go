package store

import (
	"context"
	"testing"
	"time"
)

// TestSingleEntryCache_Get tests the Get method of a single-entry cache under various conditions like validity and expiration.
func TestSingleEntryCache_Get(t *testing.T) {
	t.Run("cache not valid", func(t *testing.T) {
		cache := NewSingleCache[string](time.Minute)
		value, exists, err := cache.Get(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected no value, got one")
		}
		if value != "" {
			t.Errorf("expected empty value, got: %v", value)
		}
	})

	t.Run("cache expired", func(t *testing.T) {
		cache := NewSingleCache[string](time.Second)
		_ = cache.Set(context.Background(), "key", "value")
		time.Sleep(2 * time.Second)
		value, exists, err := cache.Get(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected no value, got one")
		}
		if value != "" {
			t.Errorf("expected empty value, got: %v", value)
		}
	})

	t.Run("value exists and valid", func(t *testing.T) {
		cache := NewSingleCache[string](time.Minute)
		_ = cache.Set(context.Background(), "key", "value")
		value, exists, err := cache.Get(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected value, got none")
		}
		if value != "value" {
			t.Errorf("expected value to be 'value', got: %v", value)
		}
	})
}

// TestSingleEntryCache_Set verifies the Set method of a single-entry cache for storing and overwriting values in the cache.
func TestSingleEntryCache_Set(t *testing.T) {
	t.Run("set single value", func(t *testing.T) {
		cache := NewSingleCache[int](time.Minute)
		err := cache.Set(context.Background(), "key", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		value, exists, err := cache.Get(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected value, got none")
		}
		if value != 42 {
			t.Errorf("expected value to be 42, got: %v", value)
		}
	})

	t.Run("overwrite value", func(t *testing.T) {
		cache := NewSingleCache[string](time.Minute)
		_ = cache.Set(context.Background(), "key", "value1")
		_ = cache.Set(context.Background(), "key", "value2")
		value, exists, err := cache.Get(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected value, got none")
		}
		if value != "value2" {
			t.Errorf("expected value to be 'value2', got: %v", value)
		}
	})
}

// TestNewSingleCache validates the creation and behavior of a single-entry cache with a specified time-to-live (TTL).
func TestNewSingleCache(t *testing.T) {
	t.Run("create new cache", func(t *testing.T) {
		cache := NewSingleCache[int](time.Minute)
		if cache == nil {
			t.Fatal("expected cache to be non-nil")
		}
		value, exists, err := cache.Get(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected no value, got one")
		}
		if value != 0 {
			t.Errorf("expected default value, got: %v", value)
		}
	})
}
