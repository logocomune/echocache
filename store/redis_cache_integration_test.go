package store

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

// redisContainer represents a Redis container instance used for testing with testcontainers.
// It includes the container instance, host address, and port number.
type redisContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

// setupRedisForTest creates and starts a Redis container for testing, returning its host and port information.
func setupRedisForTest(ctx context.Context) (*redisContainer, error) {
	port := "6379"
	proto := "tcp"
	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{port + "/" + proto},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	var redisC *redisContainer
	if container != nil {
		redisC = &redisContainer{
			Container: container,
		}
	}
	if err != nil {
		return redisC, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		return redisC, err
	}
	redisC.Host = host

	natP, err := nat.NewPort(proto, port)
	if err != nil {
		return redisC, err
	}
	mappedPort, err := container.MappedPort(ctx, natP)
	if err != nil {
		return redisC, err
	}
	redisC.Port = mappedPort.Port()

	return redisC, err

}

// getRedisClientForTest creates and returns a Redis client configured for testing purposes with minimal retry attempts.
func getRedisClientForTest(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:       addr,
		ClientName: "test-client",
		MaxRetries: 1,
	})
}

// TestRedisIntegration01 verifies Redis integration by testing cache operations, including locking and releasing, with assertions.
func TestRedisIntegration01(t *testing.T) {
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

	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	cache := NewStaleWhileRevalidateRedisCache[TestStruct](client, "t1est", time.Second*2)
	lockKey := "FirstKey"
	randValue := "FirstValue"
	locked, err := cache.TryAcquireRefreshLock(ctx, lockKey, randValue, time.Second)
	assert.Nil(t, err)
	assert.True(t, locked)
	locked, err = cache.TryAcquireRefreshLock(ctx, lockKey, randValue+"changed", time.Second)
	assert.Nil(t, err)
	assert.False(t, locked)
	err = cache.ReleaseRefreshLock(ctx, lockKey, randValue)
	assert.Nil(t, err)
	locked, err = cache.TryAcquireRefreshLock(ctx, lockKey, randValue+"changed", time.Second)
	assert.Nil(t, err)
	assert.True(t, locked)
	err = cache.ReleaseRefreshLock(ctx, lockKey, randValue+"changed")
	assert.NoError(t, err)
}
