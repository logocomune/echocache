package main

import (
	"context"
	"fmt"
	"github.com/logocomune/echocache"
	"sync"
	"time"
)

// simulateComputation simulates a computation that is memoized and returns the result.
func simulateComputation(ctx context.Context, ec *echocache.EchoCache[string], key string) (string, bool, error) {
	// Define the computation logic inside the RefreshFunc
	return ec.Memoize(ctx, key, func(ctx context.Context) (string, error) {
		time.Sleep(1 * time.Second) // Simulate expensive computation
		return "test1", nil
	})
}

// main demonstrates usage of EchoCache with concurrent memoization.
func main() {
	// Create a new cache instance with an LRU cache of size 2
	cache := echocache.New(echocache.NewLRUCache[string](2))

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
