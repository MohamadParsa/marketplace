package bootstrap_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

func httpShippingMethod() backoffice.ShippingMethodInput {
	capacity := 23
	return backoffice.ShippingMethodInput{ID: "collect-test", Name: "پس‌کرایه آزمون", Active: true, CollectionMode: "collect", DeliveryMode: "range", DailyCapacity: &capacity, CategoryMode: "selected", AllowedCategories: []string{"regular"}, AllCities: false, CityIDs: []string{"tehran-city"}, FeeToman: "", MinDays: 2, MaxDays: 4, DayBasis: shipping.CalendarDays}
}

func TestShippingMethodHTTPPermissionsFieldsAndLegacySettingsCannotBypass(t *testing.T) {
	ctx := context.Background()
	seed := storeEligibilitySeed(t)
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "shipping-view", Name: "مشاهده حمل آزمون", Version: 1, Scope: identity.ScopeGlobal, Permissions: []string{backoffice.CommerceSettingsRead}})
	for _, user := range seed.Users {
		if user.Phone == "09000000005" {
			seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: user.ID, RoleID: "shipping-view"})
		}
	}
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	in := httpShippingMethod()
	commerceHTTP(t, h, "", "", "GET", "/backoffice/shipping/methods", nil, "", 401)
	commerceHTTP(t, h, seller, "", "POST", "/backoffice/shipping/methods", in, "seller-method", 401)
	commerceHTTP(t, h, viewer, "", "POST", "/backoffice/shipping/methods", in, "viewer-method", 403)
	var read backoffice.ShippingMethodDirectory
	json.Unmarshal(commerceHTTP(t, h, viewer, "", "GET", "/backoffice/shipping/methods", nil, "", 200), &read)
	if read.CanWrite {
		t.Fatal("viewer write")
	}
	for _, bad := range []any{map[string]any{"id": "missing-version", "name": "آزمون"}, map[string]any{"id": "bad-fraction", "version": 0, "name": "آزمون", "daily_capacity": 0.5}, map[string]any{"id": "injected", "version": 0, "name": "آزمون", "owner": "fake"}} {
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/shipping/methods", bad, "invalid-method", 400)
	}
	bad := in
	bad.FeeToman = "12"
	var failure struct {
		Code   string            `json:"code"`
		Fields map[string]string `json:"fields"`
	}
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/shipping/methods", bad, "bad-collect-fee", 400), &failure)
	if failure.Code != "invalid_shipping_method" || failure.Fields["fee_toman"] == "" {
		t.Fatal("field errors", failure)
	}
	var saved backoffice.ShippingMethodInput
	first := commerceHTTP(t, h, admin, "", "POST", "/backoffice/shipping/methods", in, "create-collect", 200)
	json.Unmarshal(first, &saved)
	if saved.Version != 1 || saved.CollectionMode != "collect" || saved.FeeToman != "" {
		t.Fatal(saved)
	}
	if string(first) != string(commerceHTTP(t, h, admin, "", "POST", "/backoffice/shipping/methods", in, "create-collect", 200)) {
		t.Fatal("retry")
	}
	commerceHTTP(t, h, admin, "", "PUT", "/backoffice/shipping/methods/collect-test", in, "stale-method", 409)
	commerceHTTP(t, h, admin, "", "PUT", "/backoffice/shipping/methods/wrong-id", saved, "wrong-id", 400)
	commerceHTTP(t, h, admin, "", "GET", "/backoffice/shipping/methods?page=invalid", nil, "", 400)
	commerceHTTP(t, h, admin, "", "GET", "/backoffice/shipping/methods?size=101", nil, "", 400)
	var settings backoffice.CommerceSettingsView
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", "/backoffice/commerce/settings", nil, "", 200), &settings)
	before, _ := repo.Snapshot(ctx)
	for index := range settings.Shipping {
		if settings.Shipping[index].ID == in.ID {
			settings.Shipping[index].FeeToman = "10"
		}
	}
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", backoffice.CommerceSettingsInput{Version: settings.Version, Shipping: &backoffice.ShippingSettingsInput{Methods: settings.Shipping, ClosedWeekdays: settings.ClosedWeekdays, Holidays: settings.Holidays}}, "legacy-settings-bypass", 400)
	after, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Commerce.Settings, after.Commerce.Settings) {
		t.Fatal("failed settings changed method")
	}
}

func TestShippingMethodPostgresRestartRetainsVersionCoverageAndCollect(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := storeEligibilitySeed(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	before, _ := repo.Snapshot(ctx)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/shipping/methods", httpShippingMethod(), "persist-method", 200)
	restarted, err := reviewpostgres.New(ctx, db, seed)
	if err != nil {
		t.Fatal(err)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted)
	admin2 := login(t, h2, "09000000001")
	var directory backoffice.ShippingMethodDirectory
	json.Unmarshal(commerceHTTP(t, h2, admin2, "", "GET", "/backoffice/shipping/methods?collection_mode=collect&city_id=tehran-city", nil, "", 200), &directory)
	if directory.Total != 1 || len(directory.Rows) != 1 || directory.Rows[0].Version != 1 || directory.Rows[0].DailyCapacity == nil || *directory.Rows[0].DailyCapacity != 23 || directory.Rows[0].DayBasis != shipping.CalendarDays || directory.Rows[0].CategoryMode != "selected" || directory.Rows[0].CityIDs[0] != "tehran-city" {
		t.Fatal("restart", directory)
	}
	after, _ := restarted.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Offers, after.Offers) {
		t.Fatal("protected data")
	}
}
