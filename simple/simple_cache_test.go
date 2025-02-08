package simple

import (
	"context"
	"sync"
	"testing"
)

func TestSimpleCache_Get(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func() *SimpleCache[int]
		expectedValue  int
		expectedExists bool
		expectedErr    error
	}{
		{
			name: "nil value",
			setupFunc: func() *SimpleCache[int] {
				return NewSimpleCache[int]()
			},
			expectedValue:  0,
			expectedExists: false,
			expectedErr:    nil,
		},
		{
			name: "non-nil value",
			setupFunc: func() *SimpleCache[int] {
				c := NewSimpleCache[int]()
				c.Set(context.Background(), "key", 42)
				return c
			},
			expectedValue:  42,
			expectedExists: true,
			expectedErr:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.setupFunc()
			value, exists, err := cache.Get(context.Background(), "key")

			if value != tc.expectedValue {
				t.Errorf("expected value %v, got %v", tc.expectedValue, value)
			}
			if exists != tc.expectedExists {
				t.Errorf("expected exists %v, got %v", tc.expectedExists, exists)
			}
			if err != tc.expectedErr {
				t.Errorf("expected err %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestSimpleCache_Set(t *testing.T) {
	tests := []struct {
		name        string
		valueToSet  int
		expectedErr error
		verifyFunc  func(t *testing.T, cache *SimpleCache[int])
	}{
		{
			name:        "set value",
			valueToSet:  42,
			expectedErr: nil,
			verifyFunc: func(t *testing.T, cache *SimpleCache[int]) {
				value, exists, err := cache.Get(context.Background(), "key")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !exists {
					t.Errorf("expected value to exist, but it does not")
				}
				if value != 42 {
					t.Errorf("expected value %v, got %v", 42, value)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewSimpleCache[int]()
			err := cache.Set(context.Background(), "key", tc.valueToSet)

			if err != tc.expectedErr {
				t.Errorf("expected err %v, got %v", tc.expectedErr, err)
			}

			tc.verifyFunc(t, cache)
		})
	}
}

func TestSimpleCache_ConcurrentAccess(t *testing.T) {
	cache := NewSimpleCache[int]()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Set and Get in parallel
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := cache.Set(ctx, "key", 99)
		if err != nil {
			t.Errorf("unexpected Set error: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		_, _, err := cache.Get(ctx, "key")
		if err != nil {
			t.Errorf("unexpected Get error: %v", err)
		}
	}()
	wg.Wait()

	// Verify final value
	value, exists, err := cache.Get(ctx, "key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected value to exist, but it does not")
	}
	if value != 99 {
		t.Errorf("expected value %v, got %v", 99, value)
	}
}

func TestNewSimpleCache(t *testing.T) {
	cache := NewSimpleCache[int]()
	if cache == nil {
		t.Errorf("expected non-nil cache but got nil")
	}
}
