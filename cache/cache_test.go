package cache

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestCache(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	client, teardown := New(true)
	defer teardown(ctx)

	const str = "TEST_ONLY"
	err := client.Set(ctx, redisKeyTraffic, str, 0).Err()
	if err != nil {
		t.Fatal(err)
	}
	cacheData, err := client.Get(ctx, redisKeyTraffic).Result()
	if err != nil {
		t.Fatal(err)
	}

	if cacheData != str {
		t.Fatal("data did not match")
	}
}
