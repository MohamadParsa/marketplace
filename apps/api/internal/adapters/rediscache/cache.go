package rediscache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"sync/atomic"
	"time"
)

// Store is a disposable public JSON cache; all failures fall back in the application.
type Store struct {
	client    *redis.Client
	openUntil atomic.Int64
}

func New(url string) (*Store, error) {
	opts, e := redis.ParseURL(url)
	if e != nil {
		return nil, errors.New("invalid public cache configuration")
	}
	opts.Protocol = 2
	opts.MaxRetries = -1
	opts.DialTimeout = 50 * time.Millisecond
	opts.ReadTimeout = 50 * time.Millisecond
	opts.WriteTimeout = 50 * time.Millisecond
	opts.PoolTimeout = 25 * time.Millisecond
	opts.ContextTimeoutEnabled = true
	opts.PoolSize = 8
	opts.MaxActiveConns = 8
	opts.DisableIdentity = true
	return &Store{client: redis.NewClient(opts)}, nil
}
func (s *Store) call(ctx context.Context, fn func(context.Context) error) error {
	if time.Now().UnixNano() < s.openUntil.Load() {
		return errors.New("public cache circuit open")
	}
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	err := fn(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		s.openUntil.Store(time.Now().Add(5 * time.Second).UnixNano())
	}
	return err
}
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	var data []byte
	e := s.call(ctx, func(ctx context.Context) error { var e error; data, e = s.client.Get(ctx, key).Bytes(); return e })
	if errors.Is(e, redis.Nil) {
		return nil, nil
	}
	return data, e
}
func (s *Store) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if len(data) > 2<<20 || ttl <= 0 {
		return errors.New("invalid public cache entry")
	}
	return s.call(ctx, func(ctx context.Context) error { return s.client.Set(ctx, key, data, ttl).Err() })
}
func (s *Store) Close() error { return s.client.Close() }
