package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"reflect"
	"strings"
	"testing"
)

func productManagementSeed(t *testing.T) backoffice.Workspace {
	w := commerceSeed()
	w.IconMedia[0].Data = iconPNG(t, 8)
	w.Orders = backofficememory.Fixture().Orders
	w.Attributes = []catalog.AttributeDefinition{{ID: "color", Code: "chr-axis-color", Name: "رنگ", Kind: "enum", Options: []string{"آبی", "قرمز"}, Version: 1}, {ID: "ram", Code: "chr-ram", Name: "رم", Kind: "number", Version: 1}, {ID: "unbound", Code: "chr-unbound", Name: "خصوصیت نامرتبط", Kind: "text", Version: 1}}
	w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "color", Axis: true, Filter: true, Position: 1}, {AttributeID: "ram", Card: true, Filter: true, Position: 2}}
	w.Products[0].Attributes = []catalog.AttributeValue{{AttributeID: "ram", Value: "8"}}
	w.Products[0].EnglishName = "Test Product"
	w.Variants[0].Values = []catalog.AttributeValue{{AttributeID: "color", Value: "آبی"}}
	return w
}
func productEdit(active bool, published bool) backoffice.Command {
	return backoffice.Command{Action: "catalog.product.update", Target: "product", Version: 1, IdempotencyKey: "product-edit", Catalog: &backoffice.CatalogInput{Name: "محصول آزمون", EnglishName: "Test Product", CategoryID: "regular", Active: &active, Published: published}}
}
func TestProductManagementFourStatesAndDirectPurchase(t *testing.T) {
	for _, active := range []bool{true, false} {
		for _, published := range []bool{true, false} {
			t.Run(fmt.Sprintf("active-%v-published-%v", active, published), func(t *testing.T) {
				repo := backofficememory.New(productManagementSeed(t))
				h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
				admin := login(t, h, "09000000001")
				cmd := productEdit(active, published)
				commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 204)
				commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 204)
				w, _ := repo.Snapshot(context.Background())
				if w.Products[0].IsActive() != active || (w.Products[0].Status == "published") != published || w.Products[0].Version != 2 {
					t.Fatal("flags or retry")
				}
				d := cacheRequest(h, "/storefront/products/product", "")
				if !published {
					if d.Code != 404 {
						t.Fatal("draft disclosed")
					}
					return
				}
				if d.Code != 200 {
					t.Fatalf("PDP %d", d.Code)
				}
				var detail backoffice.ProductDetail
				json.Unmarshal(d.Body.Bytes(), &detail)
				if detail.Active != active || detail.Offers[0].Available != active || (!active && detail.Offers[0].Price != nil) {
					t.Fatal("PDP flag/price")
				}
				var listing backoffice.ListingResult
				json.Unmarshal(commerceHTTP(t, h, "", "", "GET", "/storefront/catalog?category=regular", nil, "", 200), &listing)
				if len(listing.Products) != 1 || listing.Products[0].Inactive == active || listing.Products[0].Available != active || (!active && listing.Products[0].Price != nil) {
					t.Fatal("PLP visibility")
				}
				home := backoffice.BuildHome(w)
				if !active {
					for _, r := range home.Rails {
						for _, p := range r.Products {
							if p.ID == "product" {
								t.Fatal("inactive in home")
							}
						}
					}
					if len(home.Categories) != 0 {
						t.Fatal("inactive category chip")
					}
				}
				var guest map[string]string
				json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", map[string]any{}, "", 200), &guest)
				status := 200
				if !active {
					status = 409
				}
				commerceHTTP(t, h, "", guest["cart_token"], "POST", "/storefront/cart", map[string]any{"action": "add", "offer_id": "offer", "quantity": 1, "version": 1}, "add-product", status)
			})
		}
	}
}
func TestProductManagementValidationAndScope(t *testing.T) {
	repo := backofficememory.New(productManagementSeed(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	cmd := productEdit(false, true)
	commerceHTTP(t, h, viewer, "", "POST", "/backoffice/commands", cmd, "", 403)
	commerceHTTP(t, h, seller, "", "POST", "/seller/commands", cmd, "", 403)
	for _, attrs := range [][]catalog.AttributeValue{{{AttributeID: "unbound", Value: "x"}}, {{AttributeID: "color", Value: "قرمز"}}} {
		cmd.Catalog.Attributes = &attrs
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 400)
	}
	cmd.Catalog.Attributes = nil
	cmd.Catalog.CategoryID = "gold"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 409)
	cmd.Catalog.CategoryID = "regular"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 204)
	cmd.IdempotencyKey = "stale-product"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 409)
	// Public code is allocated by the repository and never accepted as a client field.
	injected := map[string]any{"action": "catalog.product.create", "idempotency_key": "injected-code", "catalog": map[string]any{"name": "محصول جدید", "category_id": "gold", "public_code": "chr-1"}}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", injected, "", 400)
	create := backoffice.Command{Action: "catalog.product.create", IdempotencyKey: "new-product", Catalog: &backoffice.CatalogInput{Name: "محصول مستقل جدید", EnglishName: "New Product", CategoryID: "gold"}}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 204)
	w, _ := repo.Snapshot(context.Background())
	p := w.Products[len(w.Products)-1]
	if p.Status != "draft" || !p.IsActive() {
		t.Fatal("automatic code/legacy defaults")
	}
}
func TestProductActivePostgresRestartCacheAndLegacyData(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := productManagementSeed(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	admin := login(t, h, "09000000001")
	old := cacheRequest(h, "/storefront/products/product", "")
	list := cacheRequest(h, "/storefront/catalog?category=regular", "")
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", productEdit(false, true), "", 204)
	fresh := cacheRequest(h, "/storefront/products/product", old.Header().Get("ETag"))
	if fresh.Code != 200 || fresh.Header().Get("ETag") == old.Header().Get("ETag") || !strings.Contains(fresh.Body.String(), `"active":false`) {
		t.Fatal("stale PDP")
	}
	changed := cacheRequest(h, "/storefront/catalog?category=regular", list.Header().Get("ETag"))
	if changed.Code != 200 || !strings.Contains(changed.Body.String(), `"inactive":true`) {
		t.Fatal("stale PLP")
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := reopened.Snapshot(ctx)
	if after.Products[0].IsActive() {
		t.Fatal("restart flag")
	}
	for name, pair := range map[string][2]any{"offers": {before.Offers, after.Offers}, "variants": {before.Variants, after.Variants}, "orders": {before.Orders, after.Orders}, "users": {before.Users, after.Users}, "access": {before.Access, after.Access}, "commerce": {before.Commerce, after.Commerce}, "ledger": {before.Ledger, after.Ledger}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("changed " + name)
		}
	}
	create := backoffice.Command{Action: "catalog.product.create", IdempotencyKey: "pg-new-product", Catalog: &backoffice.CatalogInput{Name: "محصول دارای کد", CategoryID: "gold"}}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 204)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 204)
	current, _ := repo.Snapshot(ctx)
	created := current.Products[len(current.Products)-1]
	if !strings.HasPrefix(created.PublicCode, "chr-") {
		t.Fatal("code not allocated")
	}
	again, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	restored, _ := again.Snapshot(ctx)
	if restored.Products[len(restored.Products)-1].PublicCode != created.PublicCode {
		t.Fatal("code changed at restart")
	}
	raw, _ := json.Marshal(catalog.Product{ID: "legacy"})
	if strings.Contains(string(raw), `"active"`) || !(catalog.Product{}).IsActive() {
		t.Fatal("legacy compatibility")
	}
}

func TestProductInactiveExistingCartKeepsIdentityAndHidesPrice(t *testing.T) {
	repo := backofficememory.New(productManagementSeed(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	var guest map[string]string
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", map[string]any{}, "", 200), &guest)
	token := guest["cart_token"]
	commerceHTTP(t, h, "", token, "POST", "/storefront/cart", map[string]any{"action": "add", "offer_id": "offer", "quantity": 1, "version": 1}, "before-inactive", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", productEdit(false, true), "", 204)
	var cart backoffice.CartView
	json.Unmarshal(commerceHTTP(t, h, "", token, "GET", "/storefront/cart", nil, "", 200), &cart)
	if cart.CanCheckout || len(cart.Items) != 1 || cart.Items[0].ProductID != backoffice.PublicReference("p", "product") || !cart.Items[0].Inactive || cart.Items[0].Available || cart.Items[0].UnitPriceToman != "" || cart.Items[0].TotalToman != "" || cart.Items[0].PreviousPriceToman != "" {
		t.Fatal("inactive cart disclosure", cart)
	}
	commerceHTTP(t, h, "", token, "POST", "/storefront/cart", map[string]any{"action": "set", "offer_id": "offer", "quantity": 2, "version": cart.Version}, "increase-inactive", 409)
	commerceHTTP(t, h, "", token, "POST", "/storefront/cart", map[string]any{"action": "remove", "offer_id": "offer", "version": cart.Version}, "remove-inactive", 200)
}

func TestInactiveDraftPublicationIsIndependentOfSalesSwitch(t *testing.T) {
	seed := productManagementSeed(t)
	seed.Products[0].Status = "draft"
	inactive := false
	seed.Products[0].Active = &inactive
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", productEdit(false, true), "", 204)
	w, _ := repo.Snapshot(context.Background())
	if w.Products[0].Status != "published" || w.Products[0].IsActive() {
		t.Fatal("active/publication coupled")
	}
}
