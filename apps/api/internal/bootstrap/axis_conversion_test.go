package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"reflect"
	"strings"
	"testing"
)

func axisConversionSeed(t *testing.T) backoffice.Workspace {
	w := commerceSeed()
	w.IconMedia[0].Data = iconPNG(t, 8)
	w.Categories[0].Version = 1
	w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "gsp-axis-color", Axis: true, Filter: true, Position: 0}}
	w.Attributes = []catalog.AttributeDefinition{{ID: "gsp-axis-color", Code: "old-color", Name: "رنگ", Kind: "enum", Options: []string{"سیاه", "سفید"}, Version: 1}, {ID: "finish", Code: "finish", Name: "پوشش", Kind: "enum", Options: []string{"مات", "براق"}, Version: 1}}
	w.Variants[0].Values = []catalog.AttributeValue{{AttributeID: "gsp-axis-color", Value: "سیاه"}}
	w.Variants = append(w.Variants, catalog.Variant{ID: "variant2", ProductID: "product", SKU: "original-sku", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "gsp-axis-color", Value: "سفید"}}})
	w.Offers = append(w.Offers, catalog.Offer{ID: "offer2", ProductID: "product", VariantID: "variant2", StoreID: "s1", Price: 2000000, Stock: 5, Version: 1})
	w.Commerce.Quotes = []commerce.Quote{{ID: "past-quote", PurchaseID: "past", Status: "confirmed", Lines: []commerce.LineSnapshot{{ProductID: "product", VariantID: "variant", VariantLabel: "سیاه", UnitPrice: 12345}}}}
	w.Orders = backofficememory.Fixture().Orders
	return w
}
func httpAxisInput() backoffice.AxisConversionInput {
	return backoffice.AxisConversionInput{CategoryID: "regular", Version: 1, AxisIDs: []string{backoffice.PublicReference("a", "finish")}, Mapping: []backoffice.AxisMapping{{VariantID: backoffice.PublicReference("v", "variant"), Values: []catalog.AttributeValue{{AttributeID: backoffice.PublicReference("a", "finish"), Value: "مات"}}}, {VariantID: backoffice.PublicReference("v", "variant2"), Values: []catalog.AttributeValue{{AttributeID: backoffice.PublicReference("a", "finish"), Value: "براق"}}}}}
}
func TestAxisHTTPReadOnlyAliasPermissionAndExplicitConfirmation(t *testing.T) {
	ctx := context.Background()
	seed := axisConversionSeed(t)
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	ops := login(t, h, "09000000002")
	seller := sellerLogin(t, h, "09120000002")
	for _, a := range []struct {
		token  string
		status int
	}{{"", 401}, {ops, 403}, {seller, 401}} {
		commerceHTTP(t, h, a.token, "", "GET", "/backoffice/categories/regular/axis-conversion", nil, "", a.status)
		commerceHTTP(t, h, a.token, "", "POST", "/backoffice/axis-conversions/preview", map[string]any{"input": httpAxisInput()}, "", a.status)
	}
	before, _ := repo.Snapshot(ctx)
	var source backoffice.AxisSource
	raw := commerceHTTP(t, h, admin, "", "GET", "/backoffice/categories/regular/axis-conversion", nil, "", 200)
	json.Unmarshal(raw, &source)
	if source.Attributes[0].ID != backoffice.PublicReference("a", "gsp-axis-color") || strings.Contains(string(raw), "gsp-axis-color") || !strings.HasPrefix(source.Variants[0].ID, "chr-v-") {
		t.Fatal("source aliases leaked")
	}
	var preview backoffice.AxisPreview
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/preview", map[string]any{"input": httpAxisInput()}, "", 200), &preview)
	after, _ := repo.Snapshot(ctx)
	if !preview.Ready || !reflect.DeepEqual(before, after) {
		t.Fatal("preview wrote or invalid", preview)
	}
	var plan backoffice.AxisPlanView
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/prepare", map[string]any{"input": httpAxisInput(), "fingerprint": preview.Fingerprint}, "http-plan", 200), &plan)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/apply", map[string]any{"plan_id": plan.ID, "fingerprint": plan.Fingerprint}, "", 400)
	req := map[string]any{"plan_id": plan.ID, "fingerprint": plan.Fingerprint, "confirm": true}
	first := commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/apply", req, "", 200)
	if string(first) != string(commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/apply", req, "", 200)) {
		t.Fatal("retry receipt changed")
	}
	after, _ = repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Offers, after.Offers) || !reflect.DeepEqual(before.Commerce.Quotes, after.Commerce.Quotes) {
		t.Fatal("history rewritten")
	}
	body := commerceHTTP(t, h, admin, "", "GET", "/backoffice/workspace", nil, "", 200)
	if strings.Contains(string(body), "axis_plans") || strings.Contains(string(body), "previous_variants") {
		t.Fatal("private plans leaked")
	}
	// Conversion permission is separately grantable, never implicit in operations or store grants.
	repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "axis-editor", Name: "تبدیل", Scope: identity.ScopeGlobal, Permissions: []string{backoffice.CategoryAxesConvert}, Version: 1})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "ops", RoleID: "axis-editor"})
		return nil
	})
	commerceHTTP(t, h, ops, "", "GET", "/backoffice/categories/regular/axis-conversion", nil, "", 200)
}
func TestAxisPostgresRollbackRestartAndPublicCacheRenewal(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = axisConversionSeed(t); return nil }); e != nil {
		t.Fatal(e)
	}
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	admin := login(t, h, "09000000001")
	old := cacheRequest(h, "/storefront/products/product", "")
	if old.Code != 200 {
		t.Fatal(old.Code, old.Body.String())
	}
	tag := old.Header().Get("ETag")
	if cacheRequest(h, "/storefront/products/product", tag).Code != 304 {
		t.Fatal("PDP cache not warmed")
	}
	listing := cacheRequest(h, "/storefront/catalog?category=regular", "")
	var preview backoffice.AxisPreview
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/preview", map[string]any{"input": httpAxisInput()}, "", 200), &preview)
	var plan backoffice.AxisPlanView
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/prepare", map[string]any{"input": httpAxisInput(), "fingerprint": preview.Fingerprint}, "pg-plan", 200), &plan)
	before, _ := repo.Snapshot(ctx)
	// Fail at the public projection, after workspace/identity writes, only in this isolated schema.
	_, e := db.Exec(ctx, `CREATE FUNCTION fail_axis_projection() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'isolated projection failure'; END $$; CREATE TRIGGER fail_axis_projection BEFORE INSERT OR UPDATE ON catalog_public_state FOR EACH ROW EXECUTE FUNCTION fail_axis_projection()`)
	if e != nil {
		t.Fatal(e)
	}
	req := map[string]any{"plan_id": plan.ID, "fingerprint": plan.Fingerprint, "confirm": true}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/apply", req, "", 503)
	failed, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, failed) {
		t.Fatal("PostgreSQL half conversion committed")
	}
	if _, e = db.Exec(ctx, "DROP TRIGGER fail_axis_projection ON catalog_public_state; DROP FUNCTION fail_axis_projection()"); e != nil {
		t.Fatal(e)
	}
	first := commerceHTTP(t, h, admin, "", "POST", "/backoffice/axis-conversions/apply", req, "", 200)
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := reopened.Snapshot(ctx)
	for name, pair := range map[string][2]any{"users": {before.Users, after.Users}, "roles": {before.Access, after.Access}, "orders": {before.Orders, after.Orders}, "offers": {before.Offers, after.Offers}, "ledger": {before.Ledger, after.Ledger}, "quotes": {before.Commerce.Quotes, after.Commerce.Quotes}, "bank and shipping": {before.Commerce.Settings, after.Commerce.Settings}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("changed", name)
		}
	}
	if after.AxisPlans[0].Receipt == nil || len(after.AxisPlans[0].PreviousVariants) != 2 {
		t.Fatal("program/history not persisted")
	}
	h2 := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, reopened, cache)
	admin2 := login(t, h2, "09000000001")
	if string(first) != string(commerceHTTP(t, h2, admin2, "", "POST", "/backoffice/axis-conversions/apply", req, "", 200)) {
		t.Fatal("restart retry changed receipt")
	}
	fresh := cacheRequest(h2, "/storefront/products/product", tag)
	if fresh.Code != 200 || fresh.Header().Get("ETag") == tag || !strings.Contains(fresh.Body.String(), "پوشش") {
		t.Fatal("PDP stale after conversion", fresh.Code)
	}
	newListing := cacheRequest(h2, "/storefront/catalog?category=regular", listing.Header().Get("ETag"))
	if newListing.Code != 200 || newListing.Header().Get("ETag") == listing.Header().Get("ETag") {
		t.Fatal("PLP cache not renewed")
	}
	public, e := reopened.PublicCatalog(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	if public.AxisPlans != nil {
		t.Fatal("conversion plan in public projection")
	}
}
