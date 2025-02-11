package echocache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/nats-io/nats.go/jetstream"
	"strings"
)

// _ ensures that NatsCache implements the Cacher interface for any type T, serving as a compile-time guarantee.
var _ Cacher[any] = &NatsCache[any]{}

// NatsCache is a generic type that provides caching functionality using NATS JetStream's Key-Value store.
// NatsCache stores data under a specific prefix and uses hashed keys to avoid collisions.
// NatsCache supports marshaling/unmarshaling values of any type `T` to/from JSON format.
type NatsCache[T any] struct {
	kv     jetstream.KeyValue
	prefix string
}

func NewNatsCache[T any](kv jetstream.KeyValue, prefix string) *NatsCache[T] {
	return &NatsCache[T]{
		kv:     kv,
		prefix: prefix,
	}
}

// Get retrieves the value associated with the specified key from the cache.
// It returns the value, a boolean indicating if the key was found, and an error, if any occurred.
func (r *NatsCache[T]) Get(ctx context.Context, k string) (T, bool, error) {
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

// Set stores a value in the cache with the specified key, marshaling it to JSON before saving to the key-value store.
// Returns an error if marshaling or saving fails.
func (r *NatsCache[T]) Set(ctx context.Context, k string, value T) error {
	key := r.buildKey(k)

	// Assuming the value can be marshalled to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.kv.Put(ctx, key, data)
	return err
}

// buildKey generates a namespaced key by combining the cache prefix with an MD5 hash of the input key.
func (r *NatsCache[T]) buildKey(key string) string {
	// Example implementation, customizable as needed

	keyHash := md5.Sum([]byte(key))
	return strings.TrimRight(r.prefix, ".") + "." + hex.EncodeToString(keyHash[:])

}
