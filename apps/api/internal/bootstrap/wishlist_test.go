package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func wishlistHTTP(t *testing.T, h http.Handler, token, method, path, body string, status int) []byte {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != status {
		t.Fatalf("wishlist %s %s: %d wanted %d: %s", method, path, rr.Code, status, rr.Body.String())
	}
	if rr.Header().Get("Cache-Control") != "no-store" || rr.Header().Get("ETag") != "" {
		t.Fatal("wishlist cached")
	}
	return rr.Body.Bytes()
}
func wishlistList(t *testing.T, b []byte) backoffice.WishlistView {
	t.Helper()
	var v backoffice.WishlistView
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func TestWishlistOwnershipIdempotenceVisibilityAndPrivacy(t *testing.T) {
	ctx := context.Background()
	w := commerceSeed()
	w.Products[0].PublicCode = "chr-1234"
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	a := loginAt(t, h, "storefront", "09000000001")["token"].(string)
	b := loginAt(t, h, "storefront", "09000000002")["token"].(string)
	endpoint := "/storefront/account/wishlist"
	wishlistHTTP(t, h, "", "GET", endpoint, "", 401)
	admin := login(t, h, "09000000001")
	wishlistHTTP(t, h, admin, "PUT", endpoint+"/product", "", 401)
	wishlistHTTP(t, h, a, "PUT", endpoint+"/product", `{"user_id":"ops"}`, 400)
	wishlistHTTP(t, h, a, "GET", endpoint+"?user_id=ops", "", 400)
	for _, q := range []string{"?page=0", "?page=-1", "?page=1000001", "?page_size=49", "?page_size=", "?page=1&page=2"} {
		wishlistHTTP(t, h, a, "GET", endpoint+q, "", 400)
	}
	for i := 0; i < 2; i++ {
		wishlistHTTP(t, h, a, "PUT", endpoint+"/product", "", 200)
	}
	v := wishlistList(t, wishlistHTTP(t, h, a, "GET", endpoint, "", 200))
	if v.Total != 1 || len(v.Items) != 1 || v.Items[0].Product == nil || v.Items[0].Product.URL != "/product/chr-1234" {
		t.Fatal(v)
	}
	if wishlistList(t, wishlistHTTP(t, h, b, "GET", endpoint, "", 200)).Total != 0 {
		t.Fatal("cross-user list")
	}
	wishlistHTTP(t, h, b, "DELETE", endpoint+"/product", "", 200)
	if wishlistList(t, wishlistHTTP(t, h, a, "GET", endpoint, "", 200)).Total != 1 {
		t.Fatal("cross-user deletion")
	}
	wishlistHTTP(t, h, b, "PUT", endpoint+"/product", "", 200)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].Stock = 0; return nil }); e != nil {
		t.Fatal(e)
	}
	v = wishlistList(t, wishlistHTTP(t, h, a, "GET", endpoint, "", 200))
	if v.Items[0].Product == nil || v.Items[0].Product.Available || v.Items[0].Product.Price != nil {
		t.Fatal("out of stock hidden or priced", v)
	}
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Products[0].Status = "draft"
		w.Products[0].Name = "private-secret-name"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	raw := wishlistHTTP(t, h, a, "GET", endpoint, "", 200)
	v = wishlistList(t, raw)
	if v.Total != 1 || v.Items[0].Product != nil || bytes.Contains(raw, []byte("private-secret-name")) || bytes.Contains(raw, []byte("image_url")) {
		t.Fatal("private product leaked", string(raw))
	}
	wishlistHTTP(t, h, a, "PUT", endpoint+"/product", "", 404)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Products = nil; return nil }); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 2; i++ {
		wishlistHTTP(t, h, a, "DELETE", endpoint+"/product", "", 200)
	}
	if wishlistList(t, wishlistHTTP(t, h, a, "GET", endpoint, "", 200)).Total != 0 || wishlistList(t, wishlistHTTP(t, h, b, "GET", endpoint, "", 200)).Total != 1 {
		t.Fatal("deleted product removal or scope")
	}
	data := productCall(t, h, "GET", "/backoffice/workspace", admin, nil, 200)
	if bytes.Contains(data.Body.Bytes(), []byte(`"wishlist"`)) {
		t.Fatal("admin workspace leaked wishlist")
	}
	state, _ := repo.Snapshot(ctx)
	projection, _ := json.Marshal(backoffice.CatalogReadWorkspace(state))
	if bytes.Contains(projection, []byte(`"wishlist"`)) {
		t.Fatal("public projection leaked wishlist")
	}
}
func TestWishlistConcurrentUniquePaginationAndRevocation(t *testing.T) {
	w := commerceSeed()
	for i := 0; i < 26; i++ {
		p := w.Products[0]
		p.ID = "extra-" + string(rune('a'+i))
		w.Products = append(w.Products, p)
	}
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	token := loginAt(t, h, "storefront", "09000000001")["token"].(string)
	endpoint := "/storefront/account/wishlist"
	var group sync.WaitGroup
	statuses := make(chan int, 20)
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			r := httptest.NewRequest("PUT", endpoint+"/product", nil)
			r.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, r)
			statuses <- rr.Code
		}()
	}
	group.Wait()
	close(statuses)
	for status := range statuses {
		if status != 200 {
			t.Fatal(status)
		}
	}
	if v := wishlistList(t, wishlistHTTP(t, h, token, "GET", endpoint, "", 200)); v.Total != 1 {
		t.Fatal("duplicates", v.Total)
	}
	for _, p := range w.Products[1:] {
		wishlistHTTP(t, h, token, "PUT", endpoint+"/"+p.ID, "", 200)
	}
	first := wishlistList(t, wishlistHTTP(t, h, token, "GET", endpoint, "", 200))
	second := wishlistList(t, wishlistHTTP(t, h, token, "GET", endpoint+"?page=2", "", 200))
	if first.Total != 27 || len(first.Items) != 24 || len(second.Items) != 3 {
		t.Fatal("pagination", first.Total, len(first.Items), len(second.Items))
	}
	ids := map[string]bool{}
	for _, p := range append(first.Items, second.Items...) {
		if ids[p.ProductID] {
			t.Fatal("duplicate page item")
		}
		ids[p.ProductID] = true
	}
	repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == "admin" {
				w.Users[i].Active = false
			}
		}
		return nil
	})
	r := httptest.NewRequest("DELETE", endpoint+"/product", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != 401 && rr.Code != 403 {
		t.Fatal("revoked actor wrote", rr.Code)
	}
	state, _ := repo.Snapshot(context.Background())
	if len(state.Wishlist) != 27 {
		t.Fatal("revocation removed bookmark")
	}
}
func TestWishlistPostgresPersistsWithoutPublicInvalidationOrFinancialChanges(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	token := loginAt(t, h, "storefront", "09000000001")["token"].(string)
	before, _ := repo.Snapshot(ctx)
	stamp, e := repo.CatalogStamp(ctx, "mobile")
	if e != nil {
		t.Fatal(e)
	}
	wishlistHTTP(t, h, token, "PUT", "/storefront/account/wishlist/mobile", "", 200)
	stampAfter, e := repo.CatalogStamp(ctx, "mobile")
	if e != nil || stamp.Generation != stampAfter.Generation || stamp.Dependencies != stampAfter.Dependencies || stamp.Product != stampAfter.Product {
		t.Fatal("private mutation invalidates public cache", stamp, stampAfter, e)
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, e := reopened.Snapshot(ctx)
	if e != nil || len(after.Wishlist) != 1 || !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Orders, after.Orders) {
		t.Fatal("restart/preservation", e)
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	token = loginAt(t, h, "storefront", "09000000001")["token"].(string)
	if v := wishlistList(t, wishlistHTTP(t, h, token, "GET", "/storefront/account/wishlist", "", 200)); v.Total != 1 || v.Items[0].Product == nil {
		t.Fatal("relogin persistence", v)
	}
	if e := reopened.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Categories {
			if w.Categories[i].ID == "mobile" {
				w.Categories[i].Published = false
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if v := wishlistList(t, wishlistHTTP(t, h, token, "GET", "/storefront/account/wishlist", "", 200)); v.Items[0].Product != nil {
		t.Fatal("ancestor publication leaked")
	}
}
