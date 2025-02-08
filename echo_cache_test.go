package echocache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockCache[T any] struct {
	data map[string]T
	mu   sync.RWMutex
	err  error
}

func (mc *mockCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if mc.err != nil {
		var zero T
		return zero, false, mc.err
	}
	value, exists := mc.data[key]
	return value, exists, nil
}

func (mc *mockCache[T]) Set(ctx context.Context, key string, value T) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.err != nil {
		return mc.err
	}
	mc.data[key] = value
	return nil
}

type mockRefreshFunc[T any] struct {
	value T
	err   error
}

func (rf *mockRefreshFunc[T]) Refresh(ctx context.Context) (T, error) {
	return rf.value, rf.err
}

func TestEchoCache_Get(t *testing.T) {
	type testCase struct {
		name         string
		cache        *mockCache[string]
		refreshFunc  *mockRefreshFunc[string]
		inputKey     string
		expectedVal  string
		expectedOk   bool
		expectedErr  error
		expectedSave bool
	}

	refreshError := errors.New("Refresh error")
	cacheGetError := errors.New("cache get error")
	tests := []testCase{
		{
			name: "value exists in cache",
			cache: &mockCache[string]{
				data: map[string]string{"key1": "cached_value"},
			},
			refreshFunc:  &mockRefreshFunc[string]{value: "new_value"},
			inputKey:     "key1",
			expectedVal:  "cached_value",
			expectedOk:   true,
			expectedErr:  nil,
			expectedSave: false,
		},
		{
			name: "value does not exist in cache, Refresh succeeds",
			cache: &mockCache[string]{
				data: map[string]string{},
			},
			refreshFunc:  &mockRefreshFunc[string]{value: "new_value"},
			inputKey:     "key2",
			expectedVal:  "new_value",
			expectedOk:   true,
			expectedErr:  nil,
			expectedSave: true,
		},
		{
			name: "value does not exist in cache, Refresh fails",
			cache: &mockCache[string]{
				data: map[string]string{},
			},
			refreshFunc: &mockRefreshFunc[string]{
				err: refreshError,
			},
			inputKey:    "key3",
			expectedVal: "",
			expectedOk:  false,
			expectedErr: refreshError,
		},
		{
			name: "cache get fails",
			cache: &mockCache[string]{
				data: map[string]string{},
				err:  cacheGetError,
			},
			refreshFunc: &mockRefreshFunc[string]{value: "irrelevant"},
			inputKey:    "key4",
			expectedVal: "",
			expectedOk:  false,
			expectedErr: cacheGetError,
		},
		{
			name: "cache set fails on Refresh",
			cache: &mockCache[string]{
				data: map[string]string{},
				err:  nil,
			},
			refreshFunc: &mockRefreshFunc[string]{value: "new_value"},
			inputKey:    "key5",
			expectedVal: "new_value",
			expectedOk:  true,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshChan := make(chan string, 1)
			cache := New[string](tt.cache, tt.refreshFunc.Refresh, time.Minute)
			cache.refreshChan = refreshChan

			ctx := context.Background()
			value, ok, err := cache.Get(ctx, tt.inputKey)

			if value != tt.expectedVal {
				t.Errorf("expected value %q, got %q", tt.expectedVal, value)
			}
			if ok != tt.expectedOk {
				t.Errorf("expected ok %v, got %v", tt.expectedOk, ok)
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedSave && len(tt.cache.data) == 0 {
				t.Errorf("expected cache to save value but it did not")
			}
		})
	}
}
