package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"reflect"
	"testing"
)

func TestSalesHTTPPermissionsLegacyRetryAndPrice(t *testing.T) {
	seed := commerceSeed()
	seed.Products[0].PublicCode = "chr-1000"
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	path := "/backoffice/commerce/sales"
	commerceHTTP(t, h, "", "", "GET", path, nil, "", 401)
	commerceHTTP(t, h, buyer, "", "GET", path, nil, "", 401)
	commerceHTTP(t, h, viewer, "", "GET", path, nil, "", 403)
	var settings backoffice.SalesDirectory
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", path, nil, "", 200), &settings)
	in := settings.SalesSettings
	in.Methods[0].MarkupBPS = 400
	commerceHTTP(t, h, viewer, "", "PUT", path, in, "forbidden-sale", 403)
	commerceHTTP(t, h, admin, "", "PUT", path, map[string]any{"version": in.Version, "methods": in.Methods, "overrides": []map[string]any{{"category_id": "regular", "method_id": "cash"}}}, "missing-rate", 400)
	commerceHTTP(t, h, admin, "", "PUT", path, map[string]any{"version": in.Version, "methods": in.Methods, "overrides": []map[string]any{{"category_id": "regular", "method_id": "cash", "markup_bps": nil}}}, "null-rate", 400)
	raw := commerceHTTP(t, h, admin, "", "PUT", path, in, "sales-save", 200)
	again := commerceHTTP(t, h, admin, "", "PUT", path, in, "sales-save", 200)
	if string(raw) != string(again) {
		t.Fatal("idempotency")
	}
	commerceHTTP(t, h, admin, "", "PUT", path, in, "stale-sale", 409)
	for _, token := range []string{"product", "chr-1000"} {
		var prices []map[string]any
		json.Unmarshal(commerceHTTP(t, h, "", "", "GET", "/storefront/products/"+token+"/sales", nil, "", 200), &prices)
		if len(prices) != 1 || prices[0]["price_toman"] != "104000" {
			t.Fatal(prices)
		}
	}
}
func TestGridExportScopeAndRevocation(t *testing.T) {
	seed := commerceSeed()
	var viewerID string
	for _, u := range seed.Users {
		if u.Phone == "09000000005" {
			viewerID = u.ID
		}
	}
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "export-review", Name: "خروجی آزمون", Scope: identity.ScopeGlobal, Permissions: []string{backoffice.GridExport}, Version: 1})
	seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: viewerID, RoleID: "export-review"})
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	token := login(t, h, "09000000005")
	admin := login(t, h, "09000000001")
	commerceHTTP(t, h, "", "", "POST", "/backoffice/grids/export-authorize", nil, "", 401)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/grids/export-authorize", nil, "", 204)
	commerceHTTP(t, h, token, "", "POST", "/backoffice/grids/export-authorize", nil, "", 204)
	before, _ := repo.Snapshot(context.Background())
	repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i, g := range w.Access.Grants {
			if g.UserID == viewerID && g.RoleID == "export-review" {
				w.Access.Grants = append(w.Access.Grants[:i], w.Access.Grants[i+1:]...)
				break
			}
		}
		return nil
	})
	commerceHTTP(t, h, token, "", "POST", "/backoffice/grids/export-authorize", nil, "", 403)
	after, _ := repo.Snapshot(context.Background())
	if !reflect.DeepEqual(before.Commerce.Entries, after.Commerce.Entries) {
		t.Fatal("export wrote ledger")
	}
}
func TestSalesPostgresRestartProjectionAndPrivateState(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := commerceSeed()
	seed.IconMedia[0].Data = iconPNG(t, 8)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/grids/export-authorize", nil, "", 204)
	stamp, e := repo.CatalogStamp(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	var d backoffice.SalesDirectory
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", "/backoffice/commerce/sales", nil, "", 200), &d)
	d.Methods[0].MarkupBPS = 400
	commerceHTTP(t, h, admin, "", "PUT", "/backoffice/commerce/sales", d.SalesSettings, "pg-sales", 200)
	after, e := repo.CatalogStamp(ctx, "")
	if e != nil || after.Dependencies <= stamp.Dependencies {
		t.Fatal("sales cache dependency unchanged", e)
	}
	next, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	w, e := next.Snapshot(ctx)
	if e != nil || w.Commerce.Sales.Methods[0].MarkupBPS != 400 {
		t.Fatal("persistence", e)
	}
	public, e := next.PublicCatalog(ctx, "product")
	if e != nil {
		t.Fatal(e)
	}
	if public.Sales == nil || public.Commerce != nil || len(public.Users) > 0 {
		t.Fatal("projection/privacy")
	}
}

func TestSellerGridExportRequiresEveryStoreGrant(t *testing.T) {
	seed := commerceSeed()
	var id string
	for _, u := range seed.Users {
		if u.Phone == "09120000002" {
			id = u.ID
		}
	}
	if id == "" {
		t.Fatal("fixture seller")
	}
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "store-export", Name: "خروجی فروشگاه", Scope: identity.ScopeStore, Permissions: []string{backoffice.GridExport}, Version: 1})
	seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: id, RoleID: "store-export", StoreID: "s1"})
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	token := loginAt(t, h, "seller", "09120000002")["token"].(string)
	commerceHTTP(t, h, token, "", "POST", "/seller/grids/export-authorize", nil, "", 204)
	repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: id, StoreID: "s2"})
		return nil
	})
	commerceHTTP(t, h, token, "", "POST", "/seller/grids/export-authorize", nil, "", 403)
}
