package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"reflect"
	"strings"
	"testing"
)

func storeEligibilitySeed(t *testing.T) backoffice.Workspace {
	w := commerceSeed()
	w.IconMedia[0].Data = iconPNG(t, 8)
	w.Sales = &commerce.SalesSettings{Version: 1, Methods: []commerce.SalesMethod{{ID: "cash", Name: "نقدی", Active: true, PaymentMethodIDs: []string{"card-transfer"}}, {ID: "credit", Name: "اعتباری آزمون", Active: true, MarkupBPS: 400, PaymentMethodIDs: []string{"local"}}}}
	return w
}
func httpStorePolicy() backoffice.StoreEligibilityInput {
	return backoffice.StoreEligibilityInput{CategoryMode: "all", AllowedCategories: []string{}, DeniedCategories: []string{}, Methods: []store.SalePermission{{MethodID: "cash", DisplayName: "نام نقدی آزمون"}, {MethodID: "credit", DisplayName: "نام اعتباری آزمون"}}}
}
func TestStoreEligibilityHTTPVersionScopesAndDraftValidation(t *testing.T) {
	repo := backofficememory.New(storeEligibilitySeed(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	in := httpStorePolicy()
	for _, who := range []struct {
		token  string
		status int
	}{{"", 401}, {viewer, 403}, {seller, 401}} {
		commerceHTTP(t, h, who.token, "", "POST", "/backoffice/stores/s1/eligibility", in, "forbidden-policy", who.status)
	}
	for _, bad := range []any{map[string]any{"category_mode": "all", "methods": []any{}}, map[string]any{"version": 0, "category_mode": "all", "methods": []any{}, "owner": "injected"}, map[string]any{"version": nil, "category_mode": "all"}} {
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/s1/eligibility", bad, "invalid-policy", 400)
	}
	var row backoffice.AdminStoreRow
	first := commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/s1/eligibility", in, "save-policy", 200)
	json.Unmarshal(first, &row)
	if row.Version != 1 || row.Eligibility.Version != 1 {
		t.Fatal("policy version")
	}
	if string(first) != string(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/s1/eligibility", in, "save-policy", 200)) {
		t.Fatal("idempotent policy response")
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/s1/eligibility", in, "stale-policy", 409)
	var d backoffice.StoreDirectory
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/query", map[string]any{"filter": backoffice.StoreFilter{SalesMethod: "credit"}}, "", 200), &d)
	if !d.CanEligibility || len(d.SalesMethods) != 2 {
		t.Fatal("policy directory capability")
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/query", map[string]any{"filter": backoffice.StoreFilter{SalesMethod: "missing"}}, "", 400)
}
func TestStoreEligibilityPostgresRestartCacheAndPreservedHistory(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := storeEligibilitySeed(t)
	seed.Stores[0].Profile = &store.Profile{Email: "private-store@example.test"}
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	admin := login(t, h, "09000000001")
	old := cacheRequest(h, "/storefront/products/product", "")
	listing := cacheRequest(h, "/storefront/catalog?category=regular", "")
	in := httpStorePolicy()
	in.DeniedCategories = []string{"regular"}
	var row backoffice.AdminStoreRow
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/s1/eligibility", in, "pg-policy", 200), &row)
	fresh := cacheRequest(h, "/storefront/products/product", old.Header().Get("ETag"))
	if fresh.Code != 200 || fresh.Header().Get("ETag") == old.Header().Get("ETag") || !strings.Contains(fresh.Body.String(), `"restricted":true`) || strings.Contains(fresh.Body.String(), "private-store@example.test") {
		t.Fatal("cache or profile disclosure")
	}
	changed := cacheRequest(h, "/storefront/catalog?category=regular", listing.Header().Get("ETag"))
	if changed.Code != 200 || changed.Header().Get("ETag") == listing.Header().Get("ETag") {
		t.Fatal("PLP stale policy")
	}
	// Existing offers stay stored; direct public add cannot bypass the denial.
	var guest map[string]string
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", map[string]any{}, "", 200), &guest)
	commerceHTTP(t, h, "", guest["cart_token"], "POST", "/storefront/cart", map[string]any{"action": "add", "offer_id": "offer", "quantity": 1, "version": 1}, "denied-cart", 409)
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := reopened.Snapshot(ctx)
	for name, pair := range map[string][2]any{"users": {before.Users, after.Users}, "roles": {before.Access, after.Access}, "offers": {before.Offers, after.Offers}, "orders": {before.Orders, after.Orders}, "ledger": {before.Ledger, after.Ledger}, "settings": {before.Commerce.Settings, after.Commerce.Settings}, "entries": {before.Commerce.Entries, after.Commerce.Entries}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("protected history changed", name)
		}
	}
	if after.Stores[0].Eligibility == nil || after.Stores[0].Eligibility.Version != 1 || after.Stores[0].Name != before.Stores[0].Name || after.Stores[0].Profile.Email != before.Stores[0].Profile.Email {
		t.Fatal("store identity/policy not persisted")
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	a2 := login(t, h2, "09000000001")
	commerceHTTP(t, h2, a2, "", "POST", "/backoffice/stores/s1/eligibility", in, "pg-policy", 200)
	public, e := reopened.PublicCatalog(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, st := range public.Stores {
		if st.ID == "s1" && (st.Eligibility == nil || st.Profile != nil) {
			t.Fatal("policy missing/private contact in catalog projection")
		}
	}
}
