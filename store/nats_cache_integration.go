package store

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/nats-io/nats.go/jetstream"
	"log/slog"
	"strings"
	"time"
)

// natsCache is a generic structure representing a cache using a NATS KeyValue store with a configurable prefix.
type natsCache[T any] struct {
	kv     jetstream.KeyValue
	prefix string
}

// NewNatsCache creates a new instance of a NATS-based cache with the specified key-value store and key prefix.
func NewNatsCache[T any](kv jetstream.KeyValue, prefix string) Cacher[T] {
	return &natsCache[T]{
		kv:     kv,
		prefix: prefix,
	}
}

// NewStaleWhileRevalidateNatsCache creates a new StaleWhileRevalidateCache instance backed by NATS JetStream KeyValue store.
// T is the type of data to be cached.
// kv specifies the KeyValue store to use for storing cached values.
// prefix defines the key prefix to use within the KeyValue store.
func NewStaleWhileRevalidateNatsCache[T any](kv jetstream.KeyValue, prefix string) StaleWhileRevalidateCache[T] {
	return &natsCache[StaleValue[T]]{
		kv:     kv,
		prefix: prefix,
	}
}

// Get retrieves the cached value for the given key. Returns the value, a boolean indicating existence, and an error.
func (r *natsCache[T]) Get(ctx context.Context, k string) (T, bool, error) {
	var emptyValue T
	key := r.buildKey(k)
	result, err := r.kv.Get(ctx, key)
	if err != nil {
		if err == jetstream.ErrKeyNotFound {
			return emptyValue, false, nil
		}
		return emptyValue, false, err
	}
	var value T

	// Assuming the value can be unmarshalled into T
	err = json.Unmarshal(result.Value(), &value)
	if err != nil {
		return emptyValue, false, err
	}
	return value, true, nil
}

// Set stores a value in the cache associated with the specified key. Returns an error if the operation fails.
func (r *natsCache[T]) Set(ctx context.Context, k string, value T) error {
	key := r.buildKey(k)

	// Assuming the value can be marshalled to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.kv.Put(ctx, key, data)
	if err != nil {
		slog.Error("Cannot set value in cache", slog.String("error", err.Error()), slog.String("cacheKey", key))
	}
	return err
}

// buildKey generates a namespaced and hashed key using the provided key and the prefix from the natsCache instance.
func (r *natsCache[T]) buildKey(key string) string {
	// Example implementation, customizable as needed

	keyHash := md5.Sum([]byte(key))
	return strings.TrimRight(r.prefix, ".") + "." + hex.EncodeToString(keyHash[:])

}

// TryAcquireRefreshLock attempts to acquire a distributed lock for a specific key with a random value and TTL.
// If the lock does not already exist, it is successfully acquired and true is returned with no error.
// If the lock exists, it checks the random value and TTL to decide whether the lock can still be acquired.
func (r *natsCache[T]) TryAcquireRefreshLock(ctx context.Context, key string, randValue string, ttl time.Duration) (bool, error) {
	lockKey := r.buildKey("lock:" + key)
	now := time.Now()

	_, err := r.kv.Create(ctx, lockKey, []byte(randValue+"|"+now.Format(time.RFC3339)))

	if err == nil {
		return true, err
	}
	if errors.Is(err, jetstream.ErrKeyExists) {
		return r.checkRandValue(ctx, key, randValue, ttl, lockKey, now)
	}
	return false, err
}

// checkRandValue verifies if the stored random value matches the given one and TTL has not expired.
// It updates refresh lock timestamp if conditions are met or attempts to reacquire lock.
func (r *natsCache[T]) checkRandValue(ctx context.Context, key string, randValue string, ttl time.Duration, lockKey string, now time.Time) (bool, error) {
	storedValue, err := r.kv.Get(ctx, lockKey)

	if err != nil || storedValue == nil {

		_ = r.kv.Delete(ctx, lockKey)
		return false, err
	}
	value := storedValue.Value()
	if value == nil {
		err = r.kv.Delete(ctx, lockKey)
		if err != nil {
			slog.Error("Cannot delete lock", slog.String("error", err.Error()), slog.String("cacheKey", lockKey))
		}
		return r.TryAcquireRefreshLock(ctx, key, randValue, ttl)
	}

	innerValues := strings.Split(string(value), "|")
	if len(innerValues) != 2 {
		err = r.kv.Delete(ctx, lockKey)
		if err != nil {
			slog.Error("Cannot delete lock", slog.String("error", err.Error()), slog.String("cacheKey", lockKey))
		}
		return r.TryAcquireRefreshLock(ctx, key, randValue, ttl)
	}
	if innerValues[0] != randValue {
		return false, nil
	}

	parse, err := time.Parse(time.RFC3339, innerValues[1])
	if err != nil {
		slog.Warn("Cannot parse lock timestamp", slog.String("error", err.Error()), slog.String("cacheKey", lockKey))
		err = r.kv.Delete(ctx, lockKey)
		if err != nil {
			slog.Error("Cannot delete lock", slog.String("error", err.Error()), slog.String("cacheKey", lockKey))
		}
		return r.TryAcquireRefreshLock(ctx, key, randValue, ttl)
	}

	if time.Since(parse) > ttl {
		err = r.kv.Delete(ctx, lockKey)
		if err != nil {
			slog.Error("Cannot delete lock", slog.String("error", err.Error()), slog.String("cacheKey", lockKey))
		}
		return r.TryAcquireRefreshLock(ctx, key, randValue, ttl)
	}
	_, err = r.kv.Put(ctx, lockKey, []byte(randValue+"|"+now.Format(time.RFC3339)))
	return true, err
}

// ReleaseRefreshLock releases the refresh lock for the given key if the supplied randValue matches the stored lock value.
func (r *natsCache[T]) ReleaseRefreshLock(ctx context.Context, key string, randValue string) error {
	lockKey := r.buildKey("lock:" + key)
	storedValue, err := r.kv.Get(ctx, lockKey)
	if err != nil {
		if err == jetstream.ErrKeyExists {
			return nil // Lock does not exist
		}
		return err
	}

	value := storedValue.Value()
	if value == nil {
		err = r.kv.Delete(ctx, lockKey)
		return err
	}

	innerValues := strings.Split(string(value), "|")

	if len(innerValues) != 2 {
		err = r.kv.Delete(ctx, lockKey)
		return err
	}
	if innerValues[0] != randValue {
		return nil
	}
	err = r.kv.Delete(ctx, lockKey)
	return err
}
