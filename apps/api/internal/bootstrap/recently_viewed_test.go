package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"reflect"
	"testing"
	"time"
)

type recentFutureClock struct{ now time.Time }

func (c recentFutureClock) Now() time.Time { return c.now }

func recentList(t *testing.T, b []byte) backoffice.RecentlyViewedView {
	t.Helper()
	var v backoffice.RecentlyViewedView
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func TestRecentlyViewedHTTPPrivacyVisibilityAndDeletion(t *testing.T) {
	ctx := context.Background()
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	a := loginAt(t, h, "storefront", "09000000001")["token"].(string)
	b := loginAt(t, h, "storefront", "09000000002")["token"].(string)
	admin := login(t, h, "09000000001")
	ep := "/storefront/account/recently-viewed"
	wishlistHTTP(t, h, "", "GET", ep, "", 401)
	wishlistHTTP(t, h, admin, "POST", ep+"/product", "", 401)
	for _, path := range []string{"?user_id=ops", "?page=0", "?page_size=49", "?page=1&page=2"} {
		wishlistHTTP(t, h, a, "GET", ep+path, "", 400)
	}
	wishlistHTTP(t, h, a, "POST", ep+"/product", `{"viewed_at":"2099","user_id":"ops"}`, 400)
	wishlistHTTP(t, h, a, "DELETE", ep+"?user_id=ops", "", 400)
	wishlistHTTP(t, h, a, "POST", ep+"/missing", "", 404)
	for i := 0; i < 3; i++ {
		wishlistHTTP(t, h, a, "POST", ep+"/product", "", 200)
	}
	v := recentList(t, wishlistHTTP(t, h, a, "GET", ep, "", 200))
	if v.Total != 1 || v.Items[0].Product == nil || v.Items[0].ViewedAt.IsZero() {
		t.Fatal(v)
	}
	if recentList(t, wishlistHTTP(t, h, b, "GET", ep, "", 200)).Total != 0 {
		t.Fatal("cross-user history")
	}
	wishlistHTTP(t, h, b, "DELETE", ep+"/product", "", 200)
	if recentList(t, wishlistHTTP(t, h, a, "GET", ep, "", 200)).Total != 1 {
		t.Fatal("cross-user removal")
	}
	wishlistHTTP(t, h, b, "POST", ep+"/product", "", 200)
	repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].Stock = 0; return nil })
	v = recentList(t, wishlistHTTP(t, h, a, "GET", ep, "", 200))
	if v.Items[0].Product == nil || v.Items[0].Product.Available || v.Items[0].Product.Price != nil {
		t.Fatal("out of stock", v)
	}
	repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Products[0].Status = "draft"
		w.Products[0].Name = "private-hidden-history-name"
		return nil
	})
	raw := wishlistHTTP(t, h, a, "GET", ep, "", 200)
	v = recentList(t, raw)
	if v.Total != 1 || v.Items[0].Product != nil || bytes.Contains(raw, []byte("private-hidden-history-name")) || bytes.Contains(raw, []byte("image_url")) {
		t.Fatal("hidden product metadata leaked", string(raw))
	}
	wishlistHTTP(t, h, a, "POST", ep+"/product", "", 404)
	raw = productCall(t, h, "GET", "/backoffice/workspace", admin, nil, 200).Body.Bytes()
	if bytes.Contains(raw, []byte(`"recently_viewed"`)) {
		t.Fatal("history in admin workspace")
	}
	state, _ := repo.Snapshot(ctx)
	public, _ := json.Marshal(backoffice.CatalogReadWorkspace(state))
	if bytes.Contains(public, []byte(`"recently_viewed"`)) {
		t.Fatal("history in public projection")
	}
	repo.Update(ctx, func(w *backoffice.Workspace) error { w.Products = nil; return nil })
	wishlistHTTP(t, h, a, "DELETE", ep+"/product", "", 200)
	wishlistHTTP(t, h, a, "DELETE", ep+"/product", "", 200)
	if recentList(t, wishlistHTTP(t, h, b, "GET", ep, "", 200)).Total != 1 {
		t.Fatal("other user's deleted-product history lost")
	}
	wishlistHTTP(t, h, b, "DELETE", ep, "", 200)
	wishlistHTTP(t, h, b, "DELETE", ep, "", 200)
	if recentList(t, wishlistHTTP(t, h, b, "GET", ep, "", 200)).Total != 0 {
		t.Fatal("clear failed")
	}
}
func TestRecentlyViewedPostgresRestartAndNoPublicInvalidation(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	token := loginAt(t, h, "storefront", "09000000001")["token"].(string)
	before, _ := repo.Snapshot(ctx)
	stamp, e := repo.CatalogStamp(ctx, "mobile")
	if e != nil {
		t.Fatal(e)
	}
	wishlistHTTP(t, h, token, "POST", "/storefront/account/recently-viewed/mobile", "", 200)
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, e := reopened.Snapshot(ctx)
	next, stampErr := reopened.CatalogStamp(ctx, "mobile")
	if e != nil || stampErr != nil || len(after.RecentlyViewed) != 1 || stamp.Generation != next.Generation || stamp.Dependencies != next.Dependencies || stamp.Product != next.Product {
		t.Fatal("persistence/cache", e, stampErr)
	}
	if !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Wishlist, after.Wishlist) {
		t.Fatal("unrelated state changed")
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	token = loginAt(t, h, "storefront", "09000000001")["token"].(string)
	if v := recentList(t, wishlistHTTP(t, h, token, "GET", "/storefront/account/recently-viewed", "", 200)); v.Total != 1 || v.Items[0].Product == nil {
		t.Fatal("relogin restart", v)
	}
	cleanup := backoffice.NewService(reopened, recentFutureClock{now: time.Now().Add(31 * 24 * time.Hour)}, authmemory.Random{})
	if n, e := cleanup.PurgeRecentlyViewed(ctx); e != nil || n != 1 {
		t.Fatal("persistent physical purge", n, e)
	}
	again, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	persisted, e := again.Snapshot(ctx)
	if e != nil || len(persisted.RecentlyViewed) != 0 {
		t.Fatal("purged data returned after reopen", e)
	}
	finalStamp, e := again.CatalogStamp(ctx, "mobile")
	if e != nil || finalStamp.Generation != stamp.Generation || finalStamp.Product != stamp.Product {
		t.Fatal("private purge invalidated public cache", e)
	}
}
