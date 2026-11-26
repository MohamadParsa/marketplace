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
	"strings"
	"testing"
)

func httpStoreShippingSeed(t *testing.T) backoffice.Workspace {
	w := storeEligibilitySeed(t)
	w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "store-shipping-test", Name: "حمل آزمون", Scope: identity.ScopeStore, Version: 1, Permissions: []string{backoffice.StoresShippingRead, backoffice.StoresShippingWrite}})
	w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "seller-test", StoreID: "s1", RoleID: "store-shipping-test"})
	return w
}

func TestStoreShippingHTTPScopedWorkflowAndStrictBodies(t *testing.T) {
	repo := backofficememory.New(httpStoreShippingSeed(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	seller := sellerLogin(t, h, "09120000002")
	viewer := login(t, h, "09000000005")
	commerceHTTP(t, h, "", "", "GET", "/backoffice/stores/shipping", nil, "", 401)
	commerceHTTP(t, h, viewer, "", "GET", "/backoffice/stores/shipping", nil, "", 403)
	commerceHTTP(t, h, seller, "", "GET", "/backoffice/stores/shipping", nil, "", 401)
	var scoped backoffice.StoreShippingDirectory
	json.Unmarshal(commerceHTTP(t, h, seller, "", "GET", "/seller/stores/shipping", nil, "", 200), &scoped)
	if len(scoped.Stores) != 1 || scoped.Stores[0].ID != "s1" || scoped.Stores[0].CanReview {
		t.Fatal(scoped)
	}
	for _, body := range []any{map[string]any{"action": "public", "method_id": "post", "stores": []any{map[string]any{"id": "s1", "version": 0}}}, map[string]any{"action": "public", "method_id": "post", "base_version": 0, "stores": []any{map[string]any{"id": "s1"}}}, map[string]any{"action": "public", "method_id": "post", "base_version": 0, "stores": []any{map[string]any{"id": "s1", "version": 0}}, "owner": "injected"}} {
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", body, "bad", 400)
	}
	method := httpShippingMethod()
	method.ID = "post"
	method.Name = "پست"
	method.CollectionMode = "prepaid"
	method.FeeToman = "37"
	method.AllCities = true
	method.CityIDs = nil
	method.CategoryMode = "all"
	method.AllowedCategories = nil
	in := backoffice.StoreShippingCommand{Action: "draft", MethodID: "post", Stores: []backoffice.StoreShippingTarget{{ID: "s1"}}, Method: &method}
	other := in
	other.Stores = []backoffice.StoreShippingTarget{{ID: "s2"}}
	commerceHTTP(t, h, seller, "", "POST", "/seller/stores/shipping", other, "other", 403)
	commerceHTTP(t, h, seller, "", "POST", "/seller/stores/shipping", in, "draft", 200)
	in.Action = "submit"
	in.Method = nil
	in.Stores[0].Version = 1
	commerceHTTP(t, h, seller, "", "POST", "/seller/stores/shipping", in, "submit", 200)
	in.Action = "approve"
	in.Stores[0].Version = 2
	in.Reason = "تفاوت هزینه و مدت بررسی شد"
	commerceHTTP(t, h, seller, "", "POST", "/seller/stores/shipping", in, "seller-review", 403)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", in, "approve", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", in, "approve", 200)
	workspace := commerceHTTP(t, h, viewer, "", "GET", "/backoffice/workspace", nil, "", 200)
	if strings.Contains(string(workspace), "stores.shipping.approve") || strings.Contains(string(workspace), "store_shipping") || strings.Contains(string(workspace), in.Reason) {
		t.Fatal("private shipping revisions leaked into generic workspace")
	}
	in.Action = "reject"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", in, "stale-review", 409)
}

func TestStoreShippingPostgresRestartPreservesPrivateRevisionsAndFinancialHistory(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := httpStoreShippingSeed(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	before, _ := repo.Snapshot(ctx)
	method := httpShippingMethod()
	method.ID = "post"
	method.Name = "پست"
	method.CollectionMode = "prepaid"
	method.FeeToman = "77"
	in := backoffice.StoreShippingCommand{Action: "draft", MethodID: "post", Stores: []backoffice.StoreShippingTarget{{ID: "s1"}}, Method: &method}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", in, "draft", 200)
	in.Action = "submit"
	in.Method = nil
	in.Stores[0].Version = 1
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", in, "submit", 200)
	in.Action = "approve"
	in.Reason = "تأیید آزمون"
	in.Stores[0].Version = 2
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/shipping", in, "approve", 200)
	restarted, err := reviewpostgres.New(ctx, db, seed)
	if err != nil {
		t.Fatal(err)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted)
	admin = login(t, h2, "09000000001")
	var rows backoffice.StoreShippingDirectory
	json.Unmarshal(commerceHTTP(t, h2, admin, "", "GET", "/backoffice/stores/shipping", nil, "", 200), &rows)
	found := false
	for _, st := range rows.Stores {
		if st.ID == "s1" {
			found = true
			if st.Version != 3 || st.Methods[0].Live == nil || st.Methods[0].Effective == nil || st.Methods[0].Effective.FeeToman != "77" {
				t.Fatal(st)
			}
		}
	}
	if !found {
		t.Fatal("store lost")
	}
	after, _ := restarted.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Commerce.Entries, after.Commerce.Entries) || !reflect.DeepEqual(before.Commerce.Settings, after.Commerce.Settings) {
		t.Fatal("protected data changed")
	}
}
