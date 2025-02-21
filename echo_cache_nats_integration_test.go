package echocache

import (
	"context"
	"github.com/logocomune/echocache/store"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"
)

// getNatsClientForTest creates and returns a NATS client connection for testing purposes using the specified host and port.
func getNatsClientForTest(host, port string) *nats.Conn {
	url := "nats://" + host + ":" + port

	nc, err := nats.Connect(url)
	if err != nil {
		panic(err)
	}

	return nc
}

// TestNatsCache_Integration verifies the integration of NATS-based caching using EchoCache with singleflight and JetStream KeyValue.
// It tests data retrieval, caching effectiveness, and interactions with NATS during key-value storage operations.
func TestNatsCache_Integration(t *testing.T) {
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
	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	kv, _ := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  "TEST_01",
		TTL:     time.Minute,
		Storage: jetstream.MemoryStorage,
	})
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	const keyName = "test01"
	e := NewEchoCache[TestStruct](store.NewNatsCache[TestStruct](kv, "test.1."))
	data, present, err := e.FetchWithCache(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "John", Age: 30}, nil
	})
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)
	data2, present, err := e.FetchWithCache(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Second * 3)
		return TestStruct{Name: "John1", Age: 31}, nil
	})
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data2)

}

// TestLazyNats02Integration tests the integration of Lazy NATS caching with a simulated key-value store in a NATS JetStream setup.
// It validates lazy cache updates, background refresh tasks, and correctness of cached data during concurrent operations.
// The test includes various scenarios of data retrieval and cache refresh behavior, ensuring data consistency and functionality.
func TestLazyNats02Integration(t *testing.T) {
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
	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  "TEST_02",
		TTL:     time.Minute,
		Storage: jetstream.MemoryStorage,
	})
	if err != nil {
		panic(err)
	}

	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	keyName := "test01" + randString(12)
	e := NewLazyEchoCache[TestStruct](store.NewStaleWhileRevalidateNatsCache[TestStruct](kv, "test.fs"), time.Minute*2)

	data, present, err := e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "John", Age: 30}, nil
	}, time.Millisecond*300)
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)

	time.Sleep(time.Second * 1)
	//Request cache get old resultValue and start a request in background
	data, present, err = e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "Mark", Age: 40}, nil
	}, time.Microsecond*3)
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)

	time.Sleep(time.Second * 1)
	_, _, _ = e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "Tom", Age: 30}, nil
	}, time.Second*3)
	data, present, err = e.FetchWithLazyRefresh(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "Tom", Age: 30}, nil
	}, time.Second*3)
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "Mark", Age: 40}, data)

}

// TestLazyNats03Integration validates the lazy refresh caching mechanism using a NATS JetStream Key-Value storage for integration testing.
func TestLazyNats03Integration(t *testing.T) {
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
	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  "TEST_02",
		TTL:     time.Minute,
		Storage: jetstream.MemoryStorage,
	})
	if err != nil {
		panic(err)
	}

	e := NewLazyEchoCache[TestStruct](store.NewStaleWhileRevalidateNatsCache[TestStruct](kv, "t1est"), time.Second*30)

	commonIntegration01(t, e)
}
