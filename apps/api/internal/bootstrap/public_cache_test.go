package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type publicMemoryCache struct {
	mu     sync.Mutex
	values map[string][]byte
	fail   bool
}

func (c *publicMemoryCache) Get(ctx context.Context, k string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fail {
		return nil, errors.New("offline")
	}
	return append([]byte{}, c.values[k]...), nil
}
func (c *publicMemoryCache) Set(ctx context.Context, k string, b []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fail {
		return errors.New("offline")
	}
	c.values[k] = append([]byte{}, b...)
	return nil
}

type countingPublicRepo struct {
	*postgresrepo.Repository
	privateReads atomic.Int64
	publicReads  atomic.Int64
}

func (r *countingPublicRepo) Snapshot(ctx context.Context) (backoffice.Workspace, error) {
	r.privateReads.Add(1)
	return r.Repository.Snapshot(ctx)
}
func (r *countingPublicRepo) PublicCatalog(ctx context.Context, id string) (backoffice.Workspace, error) {
	r.publicReads.Add(1)
	return r.Repository.PublicCatalog(ctx, id)
}
func cachePostgres(t *testing.T) (*countingPublicRepo, *pgxpool.Pool) {
	t.Helper()
	dsn := os.Getenv("MARKETPLACE_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set isolated PostgreSQL URL")
	}
	ctx := context.Background()
	control, e := pgx.Connect(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	schema := "cache_test_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	safe := pgx.Identifier{schema}.Sanitize()
	if _, e = control.Exec(ctx, "CREATE SCHEMA "+safe); e != nil {
		t.Fatal(e)
	}
	opts, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	opts.ConnConfig.RuntimeParams["search_path"] = schema
	db, e := pgxpool.NewWithConfig(ctx, opts)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close(); _, _ = control.Exec(ctx, "DROP SCHEMA "+safe+" CASCADE"); control.Close(ctx) })
	if e = reviewpostgres.Migrate(ctx, db, os.DirFS(filepath.Join("..", "..", "..", "..", "db", "migrations"))); e != nil {
		t.Fatal(e)
	}
	w := homeFixture(t)
	w.IconMedia = append(w.IconMedia, catalog.IconMedia{ID: "mobile-only", Data: iconPNG(t, 8)})
	w.Products[0].ImageID = "mobile-only"
	repo, e := reviewpostgres.New(ctx, db, w)
	if e != nil {
		t.Fatal(e)
	}
	return &countingPublicRepo{Repository: repo}, db
}
func cacheRequest(h http.Handler, path, tag string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", path, nil)
	if tag != "" {
		r.Header.Set("If-None-Match", tag)
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, r)
	return res
}
func TestPublicCachePostgresCrossProcessInvalidationAndRollback(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	memory := &publicMemoryCache{values: map[string][]byte{}}
	cache := backoffice.NewPublicCache(memory)
	site := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	other, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	admin := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, other)
	token := login(t, admin, "09000000001")
	data, res := homeHTTP(t, site, "")
	if res.Code != 200 || data.Rails[0].Products[0].Price != 1000 {
		t.Fatal(res.Code, res.Body.String())
	}
	homeTag := res.Header().Get("ETag")
	for _, path := range []string{"/storefront/products/mobile", "/storefront/catalog?category=mobile"} {
		if r := cacheRequest(site, path, ""); r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
	}
	reads := repo.publicReads.Load()
	_, res = homeHTTP(t, site, homeTag)
	if res.Code != 304 || repo.publicReads.Load() != reads || repo.privateReads.Load() != 0 {
		t.Fatal("hit loaded workspace", res.Code, repo.privateReads.Load())
	}
	before, _ := repo.CatalogStamp(ctx, "mobile")
	unrelated, _ := repo.CatalogStamp(ctx, "earbuds")
	price := int64(2500)
	command := backoffice.Command{Action: "offer.price.update", Target: "mobile", Offer: &backoffice.OfferInput{Price: &price}, Version: 1, IdempotencyKey: "cache-price-change"}
	if r := productCall(t, admin, "POST", "/backoffice/commands", token, command, 204); r.Code != 204 {
		t.Fatal(r.Code, r.Body.String())
	}
	after, _ := repo.CatalogStamp(ctx, "mobile")
	otherAfter, _ := repo.CatalogStamp(ctx, "earbuds")
	if after.Product == before.Product || after.Generation == before.Generation || otherAfter.Product != unrelated.Product || after.Dependencies != before.Dependencies {
		t.Fatal("imprecise revisions", before, after, unrelated, otherAfter)
	}
	data, res = homeHTTP(t, site, homeTag)
	if res.Code != 200 || data.Rails[0].Products[0].Price != 2500 {
		t.Fatal("stale home after other process", res.Code)
	}
	if r := cacheRequest(site, "/storefront/products/mobile", ""); r.Code != 200 || !strings.Contains(r.Body.String(), `"price":2500`) {
		t.Fatal("stale product", r.Body.String())
	}
	if r := cacheRequest(site, "/storefront/catalog?category=mobile", ""); r.Code != 200 || !strings.Contains(r.Body.String(), `"price":2500`) {
		t.Fatal("stale listing", r.Body.String())
	}
	// Idempotent retry and private-only writes don't evict public data.
	if r := productCall(t, admin, "POST", "/backoffice/commands", token, command, 204); r.Code != 204 {
		t.Fatal(r.Code)
	}
	stable, _ := repo.CatalogStamp(ctx, "mobile")
	if stable != after {
		t.Fatal("retry invalidated public cache")
	}
	err := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].Price = 999; return errors.New("rollback") })
	if err == nil {
		t.Fatal("expected rollback")
	}
	stable, _ = repo.CatalogStamp(ctx, "mobile")
	if stable != after {
		t.Fatal("rollback changed stamp")
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Users[0].Name = "Updated private identity"; return nil }); e != nil {
		t.Fatal(e)
	}
	stable, _ = repo.CatalogStamp(ctx, "mobile")
	if stable != after {
		t.Fatal("private write invalidated catalog")
	}
	// Importer's transaction also refreshes the read model.
	if e = repo.UpdateCatalogWithFiles(ctx, nil, func(w *backoffice.Workspace) error { w.Offers[0].Price = 3000; return nil }); e != nil {
		t.Fatal(e)
	}
	data, res = homeHTTP(t, site, "")
	if res.Code != 200 || data.Rails[0].Products[0].Price != 3000 {
		t.Fatal("import path bypassed invalidation")
	}
	// A Redis outage returns current data through the public projection.
	memory.mu.Lock()
	memory.fail = true
	memory.mu.Unlock()
	data, res = homeHTTP(t, site, "")
	if res.Code != 200 || data.Rails[0].Products[0].Price != 3000 {
		t.Fatal("outage blocked source")
	}
	memory.mu.Lock()
	memory.fail = false
	memory.mu.Unlock()
	// No private keys/finance/SKU/exact stock can enter public cached DTOs.
	memory.mu.Lock()
	for _, b := range memory.values {
		for _, bad := range []string{"password_hash", "phone", "wallet", "ledger", "\"stock\"", "\"reserved\"", "\"sku\""} {
			if strings.Contains(string(b), bad) {
				t.Error("private cache data", bad)
			}
		}
	}
	memory.mu.Unlock()
	// Publication invalidates content and media authorization even with old ETags.
	pdp := cacheRequest(site, "/storefront/products/mobile", "")
	image := cacheRequest(site, "/catalog/icons/mobile-only", "")
	if image.Code != 200 {
		t.Fatal(image.Code)
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Categories[0].Published = false; return nil }); e != nil {
		t.Fatal(e)
	}
	if r := cacheRequest(site, "/storefront/products/mobile", pdp.Header().Get("ETag")); r.Code != 404 {
		t.Fatal("hidden product served", r.Code)
	}
	if r := cacheRequest(site, "/catalog/icons/mobile-only", image.Header().Get("ETag")); r.Code != 404 {
		t.Fatal("hidden media served", r.Code)
	}
	data, res = homeHTTP(t, site, homeTag)
	if res.Code != 200 || len(data.Rails[0].Products) != 0 {
		t.Fatal("hidden home stale")
	}
	// Reopening a repository preserves namespace and versions, not in-process cache state.
	restart, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	a, _ := repo.CatalogStamp(ctx, "mobile")
	b, _ := restart.CatalogStamp(ctx, "mobile")
	if a != b {
		t.Fatal("restart reset versions")
	}
}

type expiryCacheClock struct{ now time.Time }

func (c *expiryCacheClock) Now() time.Time { return c.now }
func TestPublicCacheReservationExpiryAndWarmup(t *testing.T) {
	repo, _ := cachePostgres(t)
	ctx := context.Background()
	clock := &expiryCacheClock{time.Now().UTC()}
	until := clock.now.Add(time.Minute)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Offers[0].Reserved = 2
		w.Commerce = &commerce.State{Quotes: []commerce.Quote{{ID: "q", Status: "reserved", ReservedUntil: until, Lines: []commerce.LineSnapshot{{OfferID: "mobile", Quantity: 2}}}}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	memory := &publicMemoryCache{values: map[string][]byte{}}
	cache := backoffice.NewPublicCache(memory)
	service := backoffice.NewService(repo, clock, authmemory.Random{})
	service.UsePublicCache(cache)
	home, e := service.PublicHome(ctx)
	if e != nil || len(home.Rails[0].Products) != 0 {
		t.Fatal("reserved stock available", e)
	}
	before, _ := repo.CatalogStamp(ctx, "mobile")
	clock.now = until.Add(time.Second)
	home, e = service.PublicHome(ctx)
	if e != nil || len(home.Rails[0].Products) != 1 {
		t.Fatal("expiry failed", e)
	}
	after, _ := repo.CatalogStamp(ctx, "mobile")
	if after.Product == before.Product || after.NextExpiry != nil {
		t.Fatal("expiry revision missing", after)
	}
	result := service.WarmPublicCache(ctx, 100)
	if result.Skipped || result.Listings != 4 || result.Products != 4 || result.Failures != 0 {
		t.Fatal("warmup", result)
	}
	stats := cache.Stats()
	if stats.Hits == 0 || repo.privateReads.Load() > 1 {
		t.Fatal("warmup no hits or repeated financial snapshots", stats, repo.privateReads.Load())
	}
	// Private/financial validation uses the source, not public cache.
	_, e = service.ValidateSelection(ctx, "mobile", backoffice.ProductSelectionInput{VariantID: "mobile", OfferID: "mobile", Quantity: 1})
	if e != nil {
		t.Fatal(e)
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].Stock = 0; return nil }); e != nil {
		t.Fatal(e)
	}
	_, e = service.ValidateSelection(ctx, "mobile", backoffice.ProductSelectionInput{VariantID: "mobile", OfferID: "mobile", Quantity: 1})
	if !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("stale selection accepted", e)
	}
	memory.mu.Lock()
	memory.fail = true
	memory.mu.Unlock()
	result = service.WarmPublicCache(ctx, 100)
	if !result.Skipped {
		t.Fatal("unavailable warmup should skip", result)
	}
	// Marshal state only for verifying no storage changes to the business schema are required.
	state, e := repo.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = json.Marshal(state); e != nil {
		t.Fatal(e)
	}
}

func TestManagedPreviewUsesCurrentAccessWithoutFullCatalog(t *testing.T) {
	repo, _ := cachePostgres(t)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	productCall(t, h, "GET", "/backoffice/catalog/media/mobile-only", admin, nil, 200)
	productCall(t, h, "GET", "/backoffice/catalog/media/mobile-only", viewer, nil, 403)
	if repo.privateReads.Load() != 0 {
		t.Fatal("image preview loaded full catalog", repo.privateReads.Load())
	}
	if e := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].Phone == "09000000001" {
				w.Users[i].Active = false
				w.Users[i].AccessVersion++
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	productCall(t, h, "GET", "/backoffice/catalog/media/mobile-only", admin, nil, 401)
}
func TestPublicReadModelBackfillDoesNotRewriteBusinessState(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	var before, after string
	query := `SELECT md5(payload::text)||(SELECT md5(jsonb_agg(to_jsonb(u) ORDER BY id)::text) FROM review_users u) FROM review_workspace WHERE singleton`
	if e := db.QueryRow(ctx, query).Scan(&before); e != nil {
		t.Fatal(e)
	}
	// Isolated fixture only: simulate first startup after adding the read-model tables.
	old, _ := repo.CatalogStamp(ctx, "mobile")
	if _, e := db.Exec(ctx, "DELETE FROM catalog_public_state"); e != nil {
		t.Fatal(e)
	}
	rebuilt, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.QueryRow(ctx, query).Scan(&after); e != nil {
		t.Fatal(e)
	}
	if before != after {
		t.Fatal("backfill rewrote business or identities")
	}
	stamp, _ := rebuilt.CatalogStamp(ctx, "mobile")
	if stamp.Namespace == old.Namespace {
		t.Fatal("rebuilt cache namespace reused")
	}
	state, e := rebuilt.PublicCatalog(ctx, "")
	if e != nil || len(state.Products) != 4 || state.Commerce != nil || len(state.Users) != 0 {
		t.Fatal("invalid public projection", e)
	}
}
