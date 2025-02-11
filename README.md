# EchoCache

EchoCache is a caching library for Go, providing various caching implementations such as LRU, Redis, and NATS JetStream.

## Features

- **LRUCache**: Implements a Least Recently Used (LRU) cache using `hashicorp/golang-lru`.
- **LRUExpirableCache**: A variant of LRU with support for item expiration.
- **SingleEntryCache**: A cache that holds a single value with TTL.
- **RedisCache**: A Redis-based cache implementation.
- **NatsCache**: A cache based on NATS JetStream for distributed storage.

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
	"sync"
	"time"
)

func simulateComputation(ctx context.Context, ec *echocache.EchoCache[string], key string) (string, bool, error) {
	return ec.Memoize(ctx, key, func(ctx context.Context) (string, error) {
		time.Sleep(2 * time.Second)
		return "test1", nil
	})
}

func main() {
	cache := echocache.New(echocache.NewLRUCache[string](2))

	start := time.Now()
	result, cached, err := simulateComputation(context.TODO(), cache, "test")
	fmt.Println(result, cached, err, time.Since(start))

	wg := sync.WaitGroup{}
	wg.Add(2)

	for i := 0; i < 2; i++ {
		go func(id int) {
			defer wg.Done()
			start := time.Now()
			result, cached, err := simulateComputation(context.TODO(), cache, "test")
			fmt.Printf("Goroutine %d: %s, Cached: %v, Error: %v, Time Elapsed: %v\n", id, result, cached, err, time.Since(start))
		}(i)
	}

	wg.Wait()
	fmt.Println("All goroutines completed")
}

```


## License

This library is distributed under the MIT license.
