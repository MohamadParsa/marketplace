package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"sync/atomic"
	"time"
)

// PublicCacheStore is deliberately not a repository for identity, checkout or money.
type PublicCacheStore interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
}
type CacheStats struct{ Hits, Misses, Builds, Errors, Retries uint64 }
type PublicCache struct {
	store                                   PublicCacheStore
	ttls                                    map[string]time.Duration
	group                                   singleflight.Group
	slots                                   chan struct{}
	hits, misses, builds, failures, retries atomic.Uint64
}

func NewPublicCache(store PublicCacheStore) *PublicCache {
	return &PublicCache{store: store, slots: make(chan struct{}, 4)}
}
func (s *Service) UsePublicCache(c *PublicCache) { s.publicCache = c }
func (c *PublicCache) Stats() CacheStats {
	return CacheStats{c.hits.Load(), c.misses.Load(), c.builds.Load(), c.failures.Load(), c.retries.Load()}
}

type cachedEnvelope struct {
	Schema int             `json:"schema"`
	Sum    string          `json:"sum"`
	Data   json.RawMessage `json:"data"`
}

func checksum(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

var errCatalogChanging = errors.New("catalog_changed_during_read")

func stampKey(st CatalogStamp, kind, id string, args any) string {
	b, _ := json.Marshal(args)
	rev := st.Generation
	if id != "" {
		rev = st.Product
	}
	return fmt.Sprintf("cherriva:public:v5:%s:%s:%s:%d:%d:%s", st.Namespace, kind, checksum([]byte(id))[:20], rev, st.Dependencies, checksum(b))
}
func sameStamp(a, b CatalogStamp, id string) bool {
	if a.Namespace != b.Namespace || a.Dependencies != b.Dependencies {
		return false
	}
	if id != "" {
		return a.Product == b.Product
	}
	return a.Generation == b.Generation
}
func cachedPublic[T any](ctx context.Context, s *Service, kind, id string, args any, ttl time.Duration, build func(context.Context) (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	c := s.publicCache
	repo, ok := s.repository.(PublicCatalogRepository)
	if c == nil || !ok {
		return build(ctx)
	}
	if value := c.ttls[kind]; value > 0 && value < ttl {
		ttl = value
	}
	for attempt := 0; attempt < 3; attempt++ {
		st, err := repo.CatalogStamp(ctx, id)
		if err != nil {
			return zero, err
		}
		if st.NextExpiry != nil && !s.clock.Now().Before(*st.NextExpiry) {
			if err = s.ReleaseExpiredCommerce(ctx); err != nil {
				return zero, err
			}
			st, err = repo.CatalogStamp(ctx, id)
			if err != nil {
				return zero, err
			}
		}
		key := stampKey(st, kind, id, args)
		if data, e := c.store.Get(ctx, key); e == nil && len(data) > 0 && len(data) <= 2<<20 {
			var env cachedEnvelope
			var value T
			if json.Unmarshal(data, &env) == nil && env.Schema == 5 && env.Sum == checksum(env.Data) && json.Unmarshal(env.Data, &value) == nil {
				c.hits.Add(1)
				return value, nil
			}
			c.failures.Add(1)
		} else if e != nil {
			c.failures.Add(1)
		}
		c.misses.Add(1)
		result := c.group.DoChan(key, func() (any, error) {
			// A late caller may have observed a miss just before an earlier fill completed.
			if data, e := c.store.Get(ctx, key); e == nil && len(data) > 0 && len(data) <= 2<<20 {
				var env cachedEnvelope
				var value T
				if json.Unmarshal(data, &env) == nil && env.Schema == 5 && env.Sum == checksum(env.Data) && json.Unmarshal(env.Data, &value) == nil {
					c.hits.Add(1)
					return value, nil
				}
			}

			select {
			case c.slots <- struct{}{}:
				defer func() { <-c.slots }()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			c.builds.Add(1)
			value, e := build(ctx)
			if e != nil {
				return nil, e
			}
			current, e := repo.CatalogStamp(ctx, id)
			if e != nil {
				return nil, e
			}
			if !sameStamp(st, current, id) {
				c.retries.Add(1)
				return nil, errCatalogChanging
			}
			data, e := json.Marshal(value)
			if e != nil {
				return nil, e
			}
			encoded, e := json.Marshal(cachedEnvelope{Schema: 5, Sum: checksum(data), Data: data})
			if e != nil {
				return nil, e
			}
			if current.NextExpiry != nil {
				remaining := current.NextExpiry.Sub(s.clock.Now())
				if remaining < ttl {
					ttl = remaining
				}
			}
			if ttl > 0 && len(encoded) <= 2<<20 {
				if e = c.store.Set(ctx, key, encoded, ttl); e != nil {
					c.failures.Add(1)
				}
			}
			return value, nil
		})
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case result := <-result:
			if errors.Is(result.Err, errCatalogChanging) {
				continue
			}
			if result.Err != nil {
				return zero, result.Err
			}
			return result.Val.(T), nil
		}
	}
	// Bounded retries never fall back to a previously cached version.
	return zero, errCatalogChanging
}

// ConfigureTTLs is called once during composition, before requests or warmup start.
// It only permits shorter values than the agreed upper bounds.
func (c *PublicCache) ConfigureTTLs(home, product, listing, suggestions time.Duration) {
	c.ttls = map[string]time.Duration{}
	for k, p := range map[string][2]time.Duration{"home": {home, 5 * time.Minute}, "banners": {home, 5 * time.Minute}, "product": {product, 10 * time.Minute}, "listing": {listing, time.Minute}, "suggestions": {suggestions, 30 * time.Second}} {
		if p[0] > 0 && p[0] <= p[1] {
			c.ttls[k] = p[0]
		}
	}
}
