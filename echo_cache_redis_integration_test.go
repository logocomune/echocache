package echocache

import (
	"context"
	"github.com/logocomune/echocache/store"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"
)

// getRedisClientForTest creates and returns a Redis client configured for testing with a specified address and minimal retries.
func getRedisClientForTest(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:       addr,
		ClientName: "test-client",
		MaxRetries: 1,
	})
}

// TestMemorizingRedisIntegration verifies the Redis-based lazy echo cache integration and its behavior under concurrent access.
func TestMemorizingRedisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	redisC, err := setupRedisForTest(ctx)
	require.NotNil(t, redisC)
	testcontainers.CleanupContainer(t, redisC.Container)
	require.NoError(t, err)
	client := getRedisClientForTest(redisC.Host + ":" + redisC.Port)

	e := NewLazyEchoCache[TestStruct](store.NewStaleWhileRevalidateRedisCache[TestStruct](client, "t1est", time.Minute*2), time.Second*30)

	commonIntegration01(t, e)

}
