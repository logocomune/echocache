package store

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

// natsContainer represents a container instance for running a NATS server in tests.
// Container provides the underlying testcontainers.Container instance.
// Host is the hostname or IP address of the NATS container.
// Port is the exposed port number mapped to the host machine.
type natsContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

// setupNatsForTest sets up a NATS test container with JetStream enabled, returning its details or an error.
func setupNatsForTest(ctx context.Context) (*natsContainer, error) {
	port := "4222"
	proto := "tcp"
	req := testcontainers.ContainerRequest{
		Image:        "nats:2.10-alpine",
		ExposedPorts: []string{port + "/" + proto},
		Cmd:          []string{"-js"},                // Abilita JetStream
		WaitingFor:   wait.ForLog("Server is ready"), // Aspetta che sia pronto
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	var redisC *natsContainer
	if container != nil {
		redisC = &natsContainer{
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

// getNatsClientForTest creates and returns a NATS client connection for testing based on the provided host and port.
func getNatsClientForTest(host, port string) *nats.Conn {
	url := "nats://" + host + ":" + port

	nc, err := nats.Connect(url)
	if err != nil {
		panic(err)
	}

	return nc
}

// getKVForTest initializes and returns a JetStream KeyValue store for testing purposes with a specified configuration.
func getKVForTest(nc *nats.Conn) jetstream.KeyValue {
	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  "TEST_01",
		TTL:     time.Hour,
		Storage: jetstream.MemoryStorage,
	})
	if err != nil {
		panic(err)
	}

	return kv
}

// TestNatsIntegration01 verifies the integration of NATS-based components and lock mechanics in a test environment.
func TestNatsIntegration01(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	natsC, err := setupNatsForTest(ctx)
	require.NotNil(t, natsC)
	testcontainers.CleanupContainer(t, natsC.Container)
	require.NoError(t, err)

	nc := getNatsClientForTest(natsC.Host, natsC.Port)
	defer nc.Drain()

	kv := getKVForTest(nc)

	cache := NewStaleWhileRevalidateNatsCache[any](kv, "test.1.")

	lockKey := "FirstKey"
	randValue := "FirstValue"
	defer cache.ReleaseRefreshLock(ctx, lockKey, randValue+"changed")
	defer cache.ReleaseRefreshLock(ctx, lockKey, randValue)

	locked, err := cache.TryAcquireRefreshLock(ctx, lockKey, randValue, 10*time.Second)
	assert.Nil(t, err)
	assert.True(t, locked)
	locked, err = cache.TryAcquireRefreshLock(ctx, lockKey, randValue+"changed", 10*time.Second)
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
