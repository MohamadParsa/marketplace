package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type cacheClock struct{ now time.Time }

func (c cacheClock) Now() time.Time { return c.now }

type cacheIDs struct{}

func (cacheIDs) Token() (string, error) { return "id", nil }

type cacheMemory struct {
	mu      sync.Mutex
	data    map[string][]byte
	ttls    map[string]time.Duration
	fail    bool
	corrupt bool
	now     time.Time
	expires map[string]time.Time
}

func newCacheMemory() *cacheMemory {
	return &cacheMemory{data: map[string][]byte{}, ttls: map[string]time.Duration{}}
}
func (c *cacheMemory) Get(ctx context.Context, k string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fail {
		return nil, errors.New("offline")
	}
	if c.corrupt {
		return []byte(`{"schema":1,"sum":"bad","data":{"price":1}}`), nil
	}
	if !c.now.IsZero() && !c.now.Before(c.expires[k]) {
		delete(c.data, k)
	}
	return append([]byte{}, c.data[k]...), nil
}
func (c *cacheMemory) Set(ctx context.Context, k string, v []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fail {
		return errors.New("offline")
	}
	c.data[k] = append([]byte{}, v...)
	c.ttls[k] = ttl
	if !c.now.IsZero() {
		if c.expires == nil {
			c.expires = map[string]time.Time{}
		}
		c.expires[k] = c.now.Add(ttl)
	}
	return nil
}

type cacheRepo struct {
	rev       atomic.Int64
	snapshots atomic.Int64
	price     atomic.Int64
}

func (r *cacheRepo) Snapshot(context.Context) (Workspace, error) {
	r.snapshots.Add(1)
	return Workspace{}, errors.New("private snapshot must not be used")
}
func (r *cacheRepo) Update(context.Context, func(*Workspace) error) error { return nil }
func (r *cacheRepo) CatalogStamp(context.Context, string) (CatalogStamp, error) {
	return CatalogStamp{Namespace: "isolated-test", Generation: r.rev.Load(), Product: r.rev.Load()}, nil
}
func (r *cacheRepo) PublicCatalog(context.Context, string) (Workspace, error) {
	return Workspace{}, nil
}
func TestPublicCacheHitCorruptionOutageAndTTL(t *testing.T) {
	ctx := context.Background()
	repo := &cacheRepo{}
	repo.rev.Store(1)
	repo.price.Store(100)
	mem := newCacheMemory()
	cache := NewPublicCache(mem)
	s := NewService(repo, cacheClock{time.Now()}, cacheIDs{})
	s.UsePublicCache(cache)
	type card struct {
		Price int64 `json:"price"`
	}
	var builds int
	read := func() (card, error) {
		return cachedPublic(ctx, s, "home", "", nil, 5*time.Minute, func(context.Context) (card, error) { builds++; return card{repo.price.Load()}, nil })
	}
	for i := 0; i < 2; i++ {
		v, e := read()
		if e != nil || v.Price != 100 {
			t.Fatal(v, e)
		}
	}
	if builds != 1 || cache.Stats().Hits != 1 || repo.snapshots.Load() != 0 {
		t.Fatal("hit recomputed or private read", cache.Stats())
	}
	for _, ttl := range mem.ttls {
		if ttl != 5*time.Minute {
			t.Fatal("wrong home TTL", ttl)
		}
	}
	repo.price.Store(200)
	repo.rev.Add(1)
	v, e := read()
	if e != nil || v.Price != 200 {
		t.Fatal("stale version", v, e)
	}
	mem.corrupt = true
	v, e = read()
	if e != nil || v.Price != 200 {
		t.Fatal("corrupt cache was trusted")
	}
	mem.corrupt = false
	mem.fail = true
	v, e = read()
	if e != nil || v.Price != 200 {
		t.Fatal("cache outage blocked DB", v, e)
	}
	if cache.Stats().Errors < 2 {
		t.Fatal("missing cache failures")
	}
	// An entry never contains a workspace; only the explicit DTO is serialized.
	for _, b := range mem.data {
		var env cachedEnvelope
		if json.Unmarshal(b, &env) != nil {
			t.Fatal("bad envelope")
		}
		var fields map[string]any
		_ = json.Unmarshal(env.Data, &fields)
		if len(fields) != 1 || fields["price"] == nil {
			t.Fatal("unexpected cache fields", fields)
		}
	}
}
func TestPublicCacheConcurrentFillRetriesChangedRevision(t *testing.T) {
	repo := &cacheRepo{}
	repo.rev.Store(1)
	repo.price.Store(100)
	cache := NewPublicCache(newCacheMemory())
	s := NewService(repo, cacheClock{time.Now()}, cacheIDs{})
	s.UsePublicCache(cache)
	started := make(chan struct{})
	proceed := make(chan struct{})
	var count atomic.Int64
	build := func(context.Context) (int64, error) {
		n := count.Add(1)
		v := repo.price.Load()
		if n == 1 {
			close(started)
			<-proceed
		}
		return v, nil
	}
	result := make(chan int64)
	go func() {
		v, e := cachedPublic(context.Background(), s, "product", "p", nil, time.Minute, build)
		if e != nil {
			result <- -1
		} else {
			result <- v
		}
	}()
	<-started
	repo.price.Store(200)
	repo.rev.Add(1)
	close(proceed)
	if v := <-result; v != 200 {
		t.Fatal("old fill won", v)
	}
	if cache.Stats().Retries != 1 {
		t.Fatal("no version recheck", cache.Stats())
	}
	for i := 0; i < 5; i++ {
		v, e := cachedPublic(context.Background(), s, "product", "p", nil, time.Minute, build)
		if e != nil || v != 200 {
			t.Fatal(v, e)
		}
	}
	if count.Load() != 2 {
		t.Fatal("cache miss after retry", count.Load())
	}
}
func TestPublicCacheSingleflightAndCancellation(t *testing.T) {
	repo := &cacheRepo{}
	repo.rev.Store(1)
	cache := NewPublicCache(newCacheMemory())
	s := NewService(repo, cacheClock{time.Now()}, cacheIDs{})
	s.UsePublicCache(cache)
	var calls atomic.Int64
	started := make(chan struct{})
	proceed := make(chan struct{})
	var wg sync.WaitGroup
	build := func(context.Context) (int, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-proceed
		return 9, nil
	}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := cachedPublic(context.Background(), s, "home", "", nil, time.Minute, build)
			if e != nil || v != 9 {
				t.Errorf("read %d %v", v, e)
			}
		}()
	}
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e := cachedPublic(ctx, s, "home", "", nil, time.Minute, build)
	if !errors.Is(e, context.Canceled) {
		t.Fatal("cancellation lost", e)
	}
	close(proceed)
	wg.Wait()
	// Late arrivals may hit Redis, but the same cold generation is built once.
	if calls.Load() != 1 {
		t.Fatal("stampede", calls.Load())
	}
}
func TestPublicCacheFailureNeverCachesErrors(t *testing.T) {
	repo := &cacheRepo{}
	cache := NewPublicCache(newCacheMemory())
	s := NewService(repo, cacheClock{time.Now()}, cacheIDs{})
	s.UsePublicCache(cache)
	for i := 0; i < 2; i++ {
		_, e := cachedPublic(context.Background(), s, "product", "hidden", nil, time.Minute, func(context.Context) (int, error) { return 0, ErrNotFound })
		if !errors.Is(e, ErrNotFound) {
			t.Fatal(e)
		}
	}
	if cache.Stats().Builds != 2 || cache.Stats().Hits != 0 {
		t.Fatal(cache.Stats())
	}
}

func TestPublicHomeExpiresAtFiveMinutesWithControlledClock(t *testing.T) {
	repo := &cacheRepo{}
	repo.rev.Store(1)
	repo.price.Store(100)
	memory := newCacheMemory()
	memory.now = time.Unix(1000, 0)
	c := NewPublicCache(memory)
	service := NewService(repo, cacheClock{memory.now}, cacheIDs{})
	service.UsePublicCache(c)
	read := func() (int64, error) {
		return cachedPublic(context.Background(), service, "home", "", nil, 5*time.Minute, func(context.Context) (int64, error) { return repo.price.Load(), nil })
	}
	if v, e := read(); e != nil || v != 100 {
		t.Fatal(v, e)
	}
	repo.price.Store(200) // Deliberately unchanged revision, to isolate TTL expiry.
	memory.mu.Lock()
	memory.now = memory.now.Add(299 * time.Second)
	memory.mu.Unlock()
	if v, e := read(); e != nil || v != 100 {
		t.Fatal("premature expiry", v, e)
	}
	memory.mu.Lock()
	memory.now = memory.now.Add(time.Second)
	memory.mu.Unlock()
	if v, e := read(); e != nil || v != 200 {
		t.Fatal("served after 300s", v, e)
	}
}
