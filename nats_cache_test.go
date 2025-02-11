package echocache

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNatsCache_Integration(t *testing.T) {
	t.Skipf("Need a running nats instance to run this test. Run: docker run -p 4222:4222 -p 8222:8222 nats:latest")
	const url = "nats://admin:test@127.0.0.1:4222"

	nc, err := nats.Connect(url)
	if err != nil {
		panic(err)
	}
	defer nc.Drain()
	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
	e := New[TestStruct](NewNatsCache[TestStruct](kv, "test.1."))
	data, present, err := e.Memoize(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Millisecond * 300)
		return TestStruct{Name: "John", Age: 30}, nil
	})
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data)
	data2, present, err := e.Memoize(context.Background(), keyName, func(ctx context.Context) (TestStruct, error) {
		time.Sleep(time.Second * 3)
		return TestStruct{Name: "John1", Age: 31}, nil
	})
	assert.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, TestStruct{Name: "John", Age: 30}, data2)

}
