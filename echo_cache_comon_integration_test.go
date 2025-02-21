package echocache

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"sync"
	"testing"
	"time"
)

// TestStruct represents a simple data structure with a name and age for use in caching or other applications.
type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// natsContainer represents a wrapper for a NATS server container with its runtime details.
// Container holds the testcontainers.Container instance for managing the container lifecycle.
// Host defines the host address of the running NATS container.
// Port specifies the port on which the NATS server is exposed.
type natsContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

// setupNatsForTest sets up and starts a NATS container for testing, enabling JetStream, and returns connection details.
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

// redisContainer represents a wrapper for a test container running a Redis instance.
// It embeds testcontainers.Container and includes custom fields Host and Port for connection details.
type redisContainer struct {
	testcontainers.Container
	Host string
	Port string
}

// setupRedisForTest initializes a Redis container for testing, returning its connection details and any errors encountered.
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

// commonIntegration01 validates the lazy cache integration by testing data retrieval, concurrent access, and lazy refresh behavior.
func commonIntegration01(t *testing.T, e *EchoCacheLazy[TestStruct]) {
	keyName := randString(10)

	//First request no data. So data is computed
	data, present, err := e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "John", Age: 30}, nil
	}, time.Second*3)
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)
	sg := sync.WaitGroup{}
	sg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer sg.Done()
			//We request wit a refresh time of 3 seconds (data must return from cache)
			data, present, err = e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
				panic("should not be called")

			}, time.Second*3)
			assert.NoError(t, err)
			assert.True(t, present)
			assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)
		}()
	}
	sg.Wait()
	time.Sleep(time.Second * 3)
	//We get again old cache but refresh is called
	data, present, err = e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		return TestStruct{Name: "Tom", Age: 32}, nil
	}, time.Second*1)
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)
	//Wait new cache was computed
	time.Sleep(time.Millisecond * 300)

	data, present, err = e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		panic("should not be called")
	}, time.Second*10)
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "Tom", Age: 32}, data)
}
