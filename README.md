# EchoCache

EchoCache is a caching library for Go that provides various cache implementations, including LRU, Redis, and NATS JetStream. It is designed to improve application performance by reducing the load on databases and external services.

## Features

- **LRUCache**: Implements a Least Recently Used (LRU) cache using `hashicorp/golang-lru`.
- **LRUExpirableCache**: A variant of LRU with support for item expiration.
- **SingleEntryCache**: A cache that holds a single value with TTL.
- **RedisCache**: Redis-based implementation with persistence and distributed management support.
- **NatsCache**: NATS JetStream-based implementation for distributed storage and asynchronous caching.
- **Stale-While-Revalidate**: Support for asynchronously reloading stale data to avoid bottlenecks.
- **Automatic concurrency handling**: Uses `singleflight` to prevent duplicate requests for the same key.

## Installation

Ensure you have Go installed, then run:

```sh
go get github.com/logocomune/echocache
```

## Usage

### Example using EchoCache with Memoization

```go
package main

import (
	"context"
	"fmt"
	"github.com/logocomune/echocache"
	"github.com/logocomune/echocache/store"
	"sync"
	"time"
)

// simulateComputation simulates a computation that is memoized and returns the result.
func simulateComputation(ctx context.Context, ec *echocache.EchoCache[string], key string) (string, bool, error) {
	// Define the computation logic inside the RefreshFunc
	return ec.FetchWithCache(ctx, key, func(ctx context.Context) (string, error) {
		time.Sleep(1 * time.Second) // Simulate expensive computation
		return "test1", nil
	})
}

// main demonstrates usage of EchoCache with concurrent memoization.
func main() {
	// Create a new cache instance with an LRU cache of size 2
	cache := echocache.NewEchoCache[string](store.NewLRUCache[string](2))

	// Single-threaded memoize usage
	start := time.Now()
	result, cached, err := simulateComputation(context.TODO(), cache, "test")
	fmt.Println(result, cached, err, time.Since(start))

	// Use a WaitGroup to synchronize goroutines
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Launch two goroutines to simulate concurrent access
	for i := 0; i < 2; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine performs the same memoized computation
			start := time.Now()
			result, cached, err := simulateComputation(context.TODO(), cache, "test")
			fmt.Printf("Goroutine %d: %s, Cached: %v, Error: %v, Time Elapsed: %v\n", id, result, cached, err, time.Since(start))
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Println("All goroutines completed")
}


```


## License

Distributed under the MIT license.
