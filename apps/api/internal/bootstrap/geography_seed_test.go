package bootstrap_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func TestBundledGeographyAdminRoundTripAndCustomerDestination(t *testing.T) {
	raw, err := os.ReadFile("../../../../data/iran-geography/iran-provinces-cities.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := backoffice.DecodeGeographySeed(raw)
	if err != nil {
		t.Fatal(err)
	}
	w := commerceSeed()
	repo := backofficememory.New(w)
	if _, err = backoffice.ApplyGeographySeed(context.Background(), repo, seed, w.Commerce.Settings.Version, time.Now()); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	var settings struct {
		Version   int                 `json:"version"`
		Geography []commerce.Province `json:"geography"`
	}
	if err = json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", "/backoffice/commerce/settings", nil, "", 200), &settings); err != nil {
		t.Fatal(err)
	}
	in := backoffice.CommerceSettingsInput{Version: settings.Version, Geography: &settings.Geography}
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", in, "save-full-geography", 200)
	viewer := login(t, h, "09000000005")
	commerceHTTP(t, h, viewer, "", "PATCH", "/backoffice/commerce/settings", in, "denied-full-geography", 403)
	var public struct {
		Geography []commerce.Province `json:"geography"`
	}
	if err = json.Unmarshal(commerceHTTP(t, h, "", "", "GET", "/storefront/commerce/config", nil, "", 200), &public); err != nil {
		t.Fatal(err)
	}
	cities := 0
	for _, p := range public.Geography {
		cities += len(p.Cities)
	}
	if len(public.Geography) != 31 || cities != 1481 {
		t.Fatalf("public directory: %d provinces, %d cities", len(public.Geography), cities)
	}
	customer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	address := backoffice.AddressInput{HouseNumber: "12", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Recipient: "گیرنده ساختگی", Phone: "09000000003", Province: "هرمزگان", City: "کیش", Street: "نشانی آزمون", PostalCode: "1234567890"}
	commerceHTTP(t, h, customer, "", "POST", "/storefront/account/addresses", address, "new-kish-destination", 200)
	address.Province = "البرز"
	body := commerceHTTP(t, h, customer, "", "POST", "/storefront/account/addresses", address, "wrong-new-province-city", 400)
	if !strings.Contains(string(body), "address_unavailable") {
		t.Fatal("city from another province was not rejected")
	}
}
