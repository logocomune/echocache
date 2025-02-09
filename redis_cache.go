package echocache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisCache is a generic type that provides caching functionalities using Redis as the backend storage.
// T specifies the type of the items to be cached, enabling type-safe operations.
// db holds the Redis client instance to interact with the Redis server.
// prefix defines a string to prepend to cache keys, useful for namespacing cache entries.
// ttl specifies the time-to-live duration for cache entries to expire automatically.
type RedisCache[T any] struct {
	db     *redis.Client
	prefix string
	ttl    time.Duration
}

// Get retrieves the value associated with the given key from the Redis cache.
// It returns the value, a boolean indicating if the key exists, and an error if one occurred.
func (r *RedisCache[T]) Get(ctx context.Context, key string) (value T, exists bool, err error) {
	result, err := r.db.Get(ctx, key).Result()
	if err == redis.Nil {
		return value, false, nil
	} else if err != nil {
		return value, false, err
	}
	// Assuming the value can be unmarshalled into T
	err = json.Unmarshal([]byte(result), &value)
	if err != nil {
		return value, false, err
	}
	return value, true, nil
}

// Set stores the given value in the Redis cache under the specified key, using the configured TTL duration.
func (r *RedisCache[T]) Set(ctx context.Context, key string, value T) error {
	// Assuming the value can be marshalled to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.db.Set(ctx, key, string(data), r.ttl).Err()
}

// BuildKey constructs a unique key by combining the cache prefix with the provided key string.
func (r *RedisCache[T]) BuildKey(key string) string {
	// Example implementation, customizable as needed
	return r.prefix + ":" + key
}
