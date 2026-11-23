package rediscache

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRedisRoundtripAndExpiry(t *testing.T) {
	url := os.Getenv("MARKETPLACE_TEST_REDIS_URL")
	if url == "" {
		t.Skip("set isolated Redis URL")
	}
	store, e := New(url)
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	ctx := context.Background()
	key := "cherriva:test:" + time.Now().Format("150405.000000000")
	defer store.client.Del(ctx, key)
	data := []byte(`{"public":true}`)
	if e = store.Set(ctx, key, data, time.Minute); e != nil {
		t.Fatal(e)
	}
	got, e := store.Get(ctx, key)
	if e != nil || string(got) != string(data) {
		t.Fatal(string(got), e)
	}
	ttl, e := store.client.PTTL(ctx, key).Result()
	if e != nil || ttl <= 0 || ttl > time.Minute {
		t.Fatal("TTL", ttl, e)
	}
	if e = store.client.PExpireAt(ctx, key, time.Now().Add(-time.Second)).Err(); e != nil {
		t.Fatal(e)
	}
	if b, e := store.Get(ctx, key); e != nil || len(b) != 0 {
		t.Fatal("expired key served", string(b), e)
	}
}
func TestRedisUnavailableHasBoundedTimeout(t *testing.T) {
	store, e := New("redis://127.0.0.1:1/0")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	start := time.Now()
	_, e = store.Get(context.Background(), "public-test")
	if e == nil || time.Since(start) > time.Second {
		t.Fatal("unbounded fallback", e)
	}
	start = time.Now()
	_, e = store.Get(context.Background(), "public-test")
	if e == nil || !strings.Contains(e.Error(), "circuit") || time.Since(start) > 50*time.Millisecond {
		t.Fatal("circuit breaker", e)
	}
	if _, e = New("secret-invalid-url"); e == nil || strings.Contains(e.Error(), "secret") {
		t.Fatal("configuration leaked")
	}
}
