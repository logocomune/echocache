package rediscache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

// Cache is a generic type used for caching data in Redis with a specified key prefix.
type Cache[T any] struct {
	rdb    *redis.Client
	prefix string
}

// New creates and initializes a new Cache instance with the given address, database index, and key prefix.
func New[T any](addr string, db int, prefix string) *Cache[T] {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &Cache[T]{
		rdb:    rdb,
		prefix: prefix,
	}
}

// buildKey generates a Redis key by appending a specified prefix to the provided key, separated by a colon.
func (c *Cache[T]) buildKey(key string) string {
	return c.prefix + ":" + key
}

// Get retrieves a value from the cache by key and returns it if found, alongside a boolean indicating existence.
// Returns an error if a retrieval or deserialization issue occurs.
func (c *Cache[T]) Get(ctx context.Context, key string) (value T, exists bool, err error) {
	val, err := c.rdb.Get(ctx, c.buildKey(key)).Result()
	if err == redis.Nil {
		return value, false, nil // Key does not exist
	} else if err != nil {
		return value, false, err // An actual error occurred
	}

	// Deserialize the value
	err = json.Unmarshal([]byte(val), &value)
	if err != nil {
		return value, false, err
	}

	return value, true, nil
}

// Set stores a value in the cache under the specified key, serializing it to JSON format before saving it to Redis.
func (c *Cache[T]) Set(ctx context.Context, key string, value T) error {
	// Serialize the value
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Save the value in Redis
	return c.rdb.Set(ctx, c.buildKey(key), data, 0).Err()
}
