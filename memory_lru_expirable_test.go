package echocache

import (
	"context"
	"testing"
	"time"
)

func TestLRUExpirableCache_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name     string
		size     int
		ttl      time.Duration
		cacheOps func(cache *LRUExpirableCache[string])
		args     args
		want     interface{}
		wantOk   bool
		wantErr  error
	}{
		{
			name: "key cacheValid",
			size: 10,
			ttl:  10 * time.Second,
			cacheOps: func(cache *LRUExpirableCache[string]) {
				cache.Set(context.Background(), "key1", "value1")
			},
			args:    args{ctx: context.Background(), key: "key1"},
			want:    "value1",
			wantOk:  true,
			wantErr: nil,
		},
		{
			name:     "key does not exist",
			size:     10,
			ttl:      10 * time.Second,
			cacheOps: func(cache *LRUExpirableCache[string]) {},
			args:     args{ctx: context.Background(), key: "key1"},
			want:     "",
			wantOk:   false,
			wantErr:  nil,
		},
		{
			name: "key expired",
			size: 10,
			ttl:  1 * time.Second,
			cacheOps: func(cache *LRUExpirableCache[string]) {
				cache.Set(context.Background(), "key1", "value1")
				time.Sleep(2 * time.Second)
			},
			args:    args{ctx: context.Background(), key: "key1"},
			want:    "",
			wantOk:  false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewLRUExplorableCache[string](tt.size, tt.ttl)
			if tt.cacheOps != nil {
				tt.cacheOps(cache)
			}
			got, gotOk, err := cache.Get(tt.args.ctx, tt.args.key)
			if got != tt.want || gotOk != tt.wantOk || err != tt.wantErr {
				t.Errorf("Get() = %v, %v, %v, want %v, %v, %v", got, gotOk, err, tt.want, tt.wantOk, tt.wantErr)
			}
		})
	}
}

func TestLRUExpirableCache_Set(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value string
	}
	tests := []struct {
		name     string
		size     int
		ttl      time.Duration
		args     args
		wantErr  error
		validate func(cache *LRUExpirableCache[string]) bool
	}{
		{
			name:    "set value successfully",
			size:    10,
			ttl:     10 * time.Second,
			args:    args{ctx: context.Background(), key: "key1", value: "value1"},
			wantErr: nil,
			validate: func(cache *LRUExpirableCache[string]) bool {
				val, exists := cache.cache.Get("key1")
				return exists && val == "value1"
			},
		},
		{
			name:    "cache reaches size limit",
			size:    1,
			ttl:     10 * time.Second,
			args:    args{ctx: context.Background(), key: "key2", value: "value2"},
			wantErr: nil,
			validate: func(cache *LRUExpirableCache[string]) bool {
				val, exists := cache.cache.Get("key2")
				return exists && val == "value2"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewLRUExplorableCache[string](tt.size, tt.ttl)
			err := cache.Set(tt.args.ctx, tt.args.key, tt.args.value)
			if err != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.validate(cache) {
				t.Errorf("validation failed")
			}
		})
	}
}

func TestNewLRUExplorableCache(t *testing.T) {
	tests := []struct {
		name string
		size int
		ttl  time.Duration
	}{
		{
			name: "basic cache creation",
			size: 10,
			ttl:  10 * time.Second,
		},
		{
			name: "zero size cache",
			size: 0,
			ttl:  10 * time.Second,
		},
		{
			name: "zero TTL",
			size: 5,
			ttl:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewLRUExplorableCache[string](tt.size, tt.ttl)
			if cache.cache == nil {
				t.Errorf("NewLRUExplorableCache() created a nil internal cache")
			}
		})
	}
}
