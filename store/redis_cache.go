package store

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

// redisCache is a generic type that implements caching functionality using Redis for storing and retrieving data.
// It requires a Redis client, a key prefix, and a time-to-live (TTL) duration for cached entries.
type redisCache[T any] struct {
	db     *redis.Client
	prefix string
	ttl    time.Duration
}

// NewRedisCache creates a new Redis-based generic cache with a specified prefix and time-to-live duration.
func NewRedisCache[T any](db *redis.Client, prefix string, ttl time.Duration) Cacher[T] {
	return &redisCache[T]{
		db:     db,
		prefix: prefix,
		ttl:    ttl,
	}
}

// NewStaleWhileRevalidateRedisCache creates a Redis-backed stale-while-revalidate cache with the specified prefix and TTL.
func NewStaleWhileRevalidateRedisCache[T any](db *redis.Client, prefix string, ttl time.Duration) StaleWhileRevalidateCache[T] {
	return &redisCache[StaleValue[T]]{
		db:     db,
		prefix: prefix,
		ttl:    ttl,
	}
}

// Get retrieves a cached value by key from Redis. It returns the value, a boolean indicating existence, and an error if any.
func (r *redisCache[T]) Get(ctx context.Context, k string) (value T, exists bool, err error) {
	var emptyValue T
	key := r.buildKey(k)
	result, err := r.db.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return emptyValue, false, nil
		}

		return emptyValue, false, err
	}

	// Assuming the value can be unmarshalled into T
	err = json.Unmarshal([]byte(result), &value)
	if err != nil {
		return emptyValue, false, err
	}
	return value, true, nil
}

// Set stores the given value in the cache using the specified key and TTL, marshaling the value to JSON format.
// Returns an error if the marshaling or Redis operation fails.
func (r *redisCache[T]) Set(ctx context.Context, k string, value T) error {
	key := r.buildKey(k)
	// Assuming the value can be marshalled to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.db.Set(ctx, key, string(data), r.ttl).Err()
}

// buildKey constructs a complete key by appending a prefix and delimiter to the input key string.
func (r *redisCache[T]) buildKey(key string) string {
	// Example implementation, customizable as needed
	return r.prefix + ":" + key
}

// TryAcquireRefreshLock attempts to acquire a refresh lock identified by the given key and random value within a TTL duration.
// Returns true if the lock is acquired, false if the lock is held by another instance, or an error if an operation fails.
func (r *redisCache[T]) TryAcquireRefreshLock(ctx context.Context, key string, randValue string, ttl time.Duration) (bool, error) {
	lockKey := r.buildKey("lock:" + key)
	result, err := r.db.SetNX(ctx, lockKey, randValue, ttl).Result()
	if err != nil {
		return false, err
	}
	if result {
		return true, nil
	}
	storedValue, err := r.db.Get(ctx, lockKey).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if storedValue != randValue {
		return false, nil
	}
	_, err = r.db.Expire(ctx, lockKey, time.Minute).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

// ReleaseRefreshLock releases a refresh lock if the current instance holds it, identified by the key and randValue.
// It checks if the stored lock value matches the provided randValue before deletion.
// Returns an error if any issues occur during the retrieval or deletion of the lock.
func (r *redisCache[T]) ReleaseRefreshLock(ctx context.Context, key string, randValue string) error {
	lockKey := r.buildKey("lock:" + key)
	storedValue, err := r.db.Get(ctx, lockKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil // Lock does not exist
		}
		return err
	}
	if storedValue != randValue {
		return nil // Current lock was not acquired by this instance
	}
	_, err = r.db.Del(ctx, lockKey).Result()
	return err
}
