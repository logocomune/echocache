package echocache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisCache_Get(t *testing.T) {
	ctx := context.TODO()
	rdb, mock := redismock.NewClientMock()
	const prefix = "test"
	cache := RedisCache[string]{db: rdb, prefix: prefix, ttl: time.Hour}

	tests := []struct {
		name          string
		key           string
		mockResponse  interface{}
		mockError     error
		expectedVal   string
		expectedErr   error
		expectedExist bool
	}{
		{
			name:          "key cacheValid",
			key:           "existing-key",
			mockResponse:  `"value1"`,
			expectedVal:   "value1",
			expectedErr:   nil,
			expectedExist: true,
		},
		{
			name:          "key does not exist",
			key:           "missing-key",
			mockError:     redis.Nil,
			expectedVal:   "",
			expectedErr:   nil,
			expectedExist: false,
		},
		{
			name:          "redis error",
			key:           "error-key",
			mockError:     errors.New("redis error"),
			expectedVal:   "",
			expectedErr:   errors.New("redis error"),
			expectedExist: false,
		},
		{
			name:          "unmarshal error",
			key:           "invalid-json-key",
			mockResponse:  "invalid json",
			expectedVal:   "",
			expectedErr:   errors.New("invalid character 'i' looking for beginning of value"),
			expectedExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockError != nil {
				mock.ExpectGet(prefix + ":" + tc.key).SetErr(tc.mockError)
			} else {
				mock.ExpectGet(prefix + ":" + tc.key).SetVal(tc.mockResponse.(string))
			}

			val, exists, err := cache.Get(ctx, tc.key)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedExist, exists)
			assert.Equal(t, tc.expectedVal, val)
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	ctx := context.TODO()
	const prefix = "test"
	rdb, mock := redismock.NewClientMock()
	cache := RedisCache[string]{db: rdb, prefix: prefix, ttl: time.Hour}

	tests := []struct {
		name        string
		key         string
		value       string
		mockError   error
		expectedErr error
	}{
		{
			name:        "successful set",
			key:         "new-key",
			value:       "value1",
			expectedErr: nil,
			mockError:   nil,
		},
		{
			name:        "redis error",
			key:         "error-key",
			value:       "value2",
			mockError:   errors.New("redis error"),
			expectedErr: errors.New("redis error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.value)
			if err == nil {
				if (tc.mockError != nil) || (tc.expectedErr != nil) {
					mock.ExpectSet(prefix+":"+tc.key, string(data), cache.ttl).SetErr(tc.mockError)
				} else {
					mock.ExpectSet(prefix+":"+tc.key, string(data), cache.ttl).SetVal("OK")
				}
			} else {
				mock.ExpectSet(prefix+":"+tc.key, string(data), cache.ttl).SetErr(err)

			}

			err = cache.Set(ctx, tc.key, tc.value)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisIntegration(t *testing.T) {
	t.Skipf("Need a running redis instance to run this test. Run: docker run -p 6379:6379 redis:6.2.6-alpine3.14")
	client := redis.NewClient(&redis.Options{
		Addr:       "127.0.0.1:6379",
		ClientName: "test-client",
		MaxRetries: 1,
	})

	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	const keyName = "test01"
	e := New[TestStruct](NewRedisCache[TestStruct](client, "t1est", time.Second*2))
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
