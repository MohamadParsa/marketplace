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
	"reflect"
	"testing"
)

func testRecipientHTTP(t *testing.T, repo backoffice.Repository) (string, commerce.Address) {
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	commerceHTTP(t, h, buyer, "", "PATCH", "/storefront/account/profile", map[string]string{"name": "خریدار ساختگی", "national_id": "1111111111"}, "", 400)
	commerceHTTP(t, h, buyer, "", "PATCH", "/storefront/account/profile", map[string]string{"name": "خریدار ساختگی", "national_id": "۰۰۱۳۵۴۷۸۹۵"}, "", 200)
	// Older callers changing email/password do not erase the national ID.
	commerceHTTP(t, h, buyer, "", "PATCH", "/storefront/account/profile", map[string]string{"name": "خریدار ساختگی تازه"}, "", 200)
	var profile map[string]any
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/me", nil, "", 200), &profile)
	if profile["national_id"] != "0013547895" {
		t.Fatal("profile national ID lost")
	}
	self := true
	lat, lng := 35.7, 51.4
	in := backoffice.AddressInput{RecipientSelf: &self, Province: "تهران", City: "تهران", Street: "نشانی ساختگی", HouseNumber: "12", PostalCode: "1234567890", Latitude: &lat, Longitude: &lng}
	bad := in
	bad.HouseNumber = ""
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", bad, "missing-house", 400)
	bad = in
	bad.City = "قم"
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", bad, "mismatched-city", 400)
	bad = in
	bad.RecipientSelf = nil
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", bad, "legacy-write", 400)
	var address commerce.Address
	raw := commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", in, "self-address", 200)
	json.Unmarshal(raw, &address)
	if address.Recipient != "خریدار ساختگی تازه" || address.RecipientNationalID != "0013547895" || address.Phone != "09000000003" {
		t.Fatal("self snapshot")
	}
	if string(raw) != string(commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", in, "self-address", 200)) {
		t.Fatal("retry")
	}
	other := loginAt(t, h, "storefront", "09000000004")["token"].(string)
	var addresses []commerce.Address
	json.Unmarshal(commerceHTTP(t, h, other, "", "GET", "/storefront/account/addresses", nil, "", 200), &addresses)
	if len(addresses) != 0 {
		t.Fatal("other customer sees address")
	}
	// No real service is implicitly activated, even in review mode.
	commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/address-map", nil, "", 200)
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/address-map/reverse", map[string]any{"latitude": 35.7, "longitude": 51.4}, "", 503)
	commerceHTTP(t, h, "", "", "GET", "/storefront/account/address-map", nil, "", 401)
	return address.UserID, address
}
func TestAddressRecipientHTTP(t *testing.T) {
	testRecipientHTTP(t, backofficememory.New(commerceSeed()))
}
func TestAddressRecipientPostgresPersistenceAndPrivateProjection(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = commerceSeed(); return nil }); e != nil {
		t.Fatal(e)
	}
	user, address := testRecipientHTTP(t, repo)
	restarted, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	w, e := restarted.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if len(w.Commerce.Addresses) != 1 || !reflect.DeepEqual(w.Commerce.Addresses[0], address) {
		t.Fatal("address lost after repository restart")
	}
	found := false
	for _, u := range w.Users {
		if u.ID == user {
			found = u.NationalID == "0013547895"
		}
	}
	if !found {
		t.Fatal("national ID not persisted")
	}
	// Repeat immutable migrations is safe and profile is loaded by SQL identity projection.
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted)
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	var p map[string]any
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/me", nil, "", 200), &p)
	if p["national_id"] != "0013547895" {
		t.Fatal("identity projection lost profile")
	}
}
