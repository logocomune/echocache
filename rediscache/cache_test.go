package rediscache

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Field string
}

func TestCache_buildKey(t *testing.T) {
	tests := []struct {
		prefix string
		key    string
		want   string
	}{
		{prefix: "test", key: "key1", want: "test:key1"},
		{prefix: "prod", key: "user42", want: "prod:user42"},
		{prefix: "", key: "admin", want: ":admin"},
	}

	for _, tt := range tests {
		c := &Cache[int]{prefix: tt.prefix}
		got := c.buildKey(tt.key)
		require.Equal(t, tt.want, got)
	}
}

func TestCache_Get(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()
	defer mock.ClearExpect()

	cache := &Cache[TestStruct]{rdb: mockRedis, prefix: "test"}

	tests := []struct {
		name     string
		setup    func()
		key      string
		wantVal  TestStruct
		wantExst bool
		wantErr  bool
	}{
		{
			name: "key exists",
			setup: func() {
				data, _ := json.Marshal(TestStruct{Field: "value1"})
				mock.ExpectGet("test:existing").SetVal(string(data))
			},
			key:      "existing",
			wantVal:  TestStruct{Field: "value1"},
			wantExst: true,
			wantErr:  false,
		},
		{
			name: "key does not exist",
			setup: func() {
				mock.ExpectGet("test:missing").RedisNil()
			},
			key:      "missing",
			wantExst: false,
			wantErr:  false,
		},
		{
			name: "invalid JSON",
			setup: func() {
				mock.ExpectGet("test:invalid").SetVal("invalid-json")
			},
			key:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt.setup()
		val, exists, err := cache.Get(ctx, tt.key)
		if tt.wantErr {
			require.NotNil(t, err, tt.name)
			continue
		}
		require.Nil(t, err, tt.name)
		require.Equal(t, tt.wantExst, exists, tt.name)
		if exists {
			require.Equal(t, tt.wantVal, val, tt.name)
		}
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_Set(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()
	defer mock.ClearExpect()

	cache := &Cache[TestStruct]{rdb: mockRedis, prefix: "test"}

	tests := []struct {
		name    string
		key     string
		value   TestStruct
		wantErr bool
	}{
		{
			name:    "successful set",
			key:     "key1",
			value:   TestStruct{Field: "value1"},
			wantErr: false,
		},
		{
			name:    "empty value",
			key:     "key2",
			value:   TestStruct{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		data, _ := json.Marshal(tt.value)
		mock.ExpectSet(cache.buildKey(tt.key), data, 0).SetVal("OK")

		err := cache.Set(ctx, tt.key, tt.value)
		if tt.wantErr {
			require.NotNil(t, err, tt.name)
			continue
		}
		require.Nil(t, err, tt.name)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}
