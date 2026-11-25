package backoffice_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"testing"
)

func TestGeographyPermissionsCascadeAndHistory(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	a := h.address(t, h.buyer)
	h.add(t, h.buyer, "offer1", 1)
	p := h.accept(t, h.buyer, h.quote(t, h.buyer, "local"))
	settings, err := h.s.ReadCommerceSettings(ctx, h.admin)
	if err != nil {
		t.Fatal(err)
	}
	directory := settings.Geography
	if _, err = h.s.SaveCommerceSettings(ctx, h.finance, backoffice.CommerceSettingsInput{Version: settings.Version, Geography: &directory}, h.key()); err != backoffice.ErrForbidden {
		t.Fatal("geography permission", err)
	}
	bad := backoffice.AddressInput{HouseNumber: "12", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Recipient: "گیرنده", Phone: "09000000111", Province: "تهران", City: "قم", Street: "نشانی ساختگی", PostalCode: "1234567890"}
	if _, err = h.s.SaveAddress(ctx, h.buyer, bad, h.key()); err != backoffice.ErrAddressUnavailable {
		t.Fatal("foreign-province city accepted", err)
	}
	directory[0].Cities = append(directory[0].Cities, commerce.City{ID: "rey", Name: "ری", Active: true})
	raw, err := h.s.SaveCommerceSettings(ctx, h.admin, backoffice.CommerceSettingsInput{Version: settings.Version, Geography: &directory}, h.key())
	settings = decodeCommerce[backoffice.CommerceSettingsView](t, raw, err)
	bad.City = "ری"
	raw, err = h.s.SaveAddress(ctx, h.buyer, bad, h.key())
	fresh := decodeCommerce[commerce.Address](t, raw, err)
	if fresh.City != "ری" {
		t.Fatal("new city not usable")
	}
	public, err := h.s.PublicCommerceConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(public["geography"].([]commerce.Province)) != 2 {
		t.Fatal(public)
	}
	directory = settings.Geography
	directory[0].Active = false
	_, err = h.s.SaveCommerceSettings(ctx, h.admin, backoffice.CommerceSettingsInput{Version: settings.Version, Geography: &directory}, h.key())
	if err != nil {
		t.Fatal(err)
	}
	public, _ = h.s.PublicCommerceConfig(ctx)
	if len(public["geography"].([]commerce.Province)) != 1 {
		t.Fatal("inactive province leaked")
	}
	if _, err = h.s.SaveAddress(ctx, h.buyer, bad, h.key()); err != backoffice.ErrAddressUnavailable {
		t.Fatal("inactive destination", err)
	}
	h.add(t, h.buyer, "offer1", 1)
	cart := h.cart(t, h.buyer)
	if _, err = h.s.PreviewCheckout(ctx, h.buyer, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: a.ID, Shipping: map[string]string{"store1": "post"}}); err != backoffice.ErrAddressUnavailable {
		t.Fatal("stale address bypassed quote validation", err)
	}
	old, err := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if err != nil || old[0].Quote.Address.City != "تهران" || old[0].Quote.TotalToman != p.Quote.TotalToman {
		t.Fatal("historical order rewritten", err)
	}
}
func TestGeographyRejectsDuplicatesAndMovingExistingCity(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	settings, _ := h.s.ReadCommerceSettings(ctx, h.admin)
	dirs := []commerce.Province{{ID: "tehran", Name: "تهران", Active: true, Cities: []commerce.City{{ID: "tehran-city", Name: "تهران", Active: true}, {ID: "qom-city", Name: "قم", Active: true}}}, {ID: "qom", Name: "قم", Active: true, Cities: []commerce.City{}}}
	if _, err := h.s.SaveCommerceSettings(ctx, h.admin, backoffice.CommerceSettingsInput{Version: settings.Version, Geography: &dirs}, h.key()); err != backoffice.ErrInvalid {
		t.Fatal("moved city accepted", err)
	}
	dirs = settings.Geography
	dirs = append(dirs, commerce.Province{ID: "duplicate-name", Name: " تهران ", Active: true})
	if _, err := h.s.SaveCommerceSettings(ctx, h.admin, backoffice.CommerceSettingsInput{Version: settings.Version, Geography: &dirs}, h.key()); err != backoffice.ErrInvalid {
		t.Fatal("duplicate normalized province", err)
	}
	current, _ := h.s.ReadCommerceSettings(ctx, h.admin)
	if current.Version != settings.Version || len(current.Geography) != 2 {
		t.Fatal("invalid update mutated directory")
	}
}
