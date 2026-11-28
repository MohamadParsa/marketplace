package backoffice_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func TestGatewayCurrencyDefaultsLegacyAndEditableExactRate(t *testing.T) {
	h := newCommerceHarness(t)
	configureGatewayKey(t, h)
	ctx := context.Background()
	dir, err := h.s.ReadGatewaySettings(ctx, h.admin)
	if err != nil {
		t.Fatal(err)
	}
	if dir.Items[0].Values["currency"] != "IRT" || dir.Items[0].Values["exchange_rate"] != "1" {
		t.Fatal("new defaults", dir.Items[0].Values)
	}
	if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.GatewaySettings = []commerce.GatewaySettings{{ID: "legacy", Version: 1, Provider: "zarinpal", Name: "قدیمی", AllCategories: true, Values: map[string]string{"currency": "IRR"}}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	before, _ := h.repo.Snapshot(ctx)
	dir, err = h.s.ReadGatewaySettings(ctx, h.admin)
	if err != nil || dir.Items[0].Values["currency"] != "IRR" || dir.Items[0].Values["exchange_rate"] != "10" {
		t.Fatal("legacy currency changed", err)
	}
	afterRead, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Commerce, afterRead.Commerce) {
		t.Fatal("read mutated state")
	}
	in := dir.Items[0].GatewaySettingsInput
	in.Values["exchange_rate"] = "۱۰.۲۵۰۰۰۰"
	in.Active, in.AllowWalletTopup = true, true
	raw, err := h.s.SaveGatewaySettings(ctx, h.admin, in, h.key())
	saved := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if saved.Values["exchange_rate"] != "10.25" || !saved.Active || !saved.AllowWalletTopup {
		t.Fatal("custom rate or switches not saved")
	}
	edit := saved.GatewaySettingsInput
	edit.Values = map[string]string{"currency": "IRT"} // older client has no rate
	raw, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	updated := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if updated.Values["currency"] != "IRT" || updated.Values["exchange_rate"] != "1" {
		t.Fatal("missing rate default")
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Commerce.Entries, after.Commerce.Entries) || !reflect.DeepEqual(before.Commerce.Settings.Methods, after.Commerce.Settings.Methods) {
		t.Fatal("settings changed money/public methods")
	}
}

func TestInvalidGatewayCurrencyRateDoesNotMutateState(t *testing.T) {
	h := newCommerceHarness(t)
	configureGatewayKey(t, h)
	ctx := context.Background()
	for _, tc := range []struct{ field, value string }{
		{"currency", "USD"}, {"exchange_rate", ""}, {"exchange_rate", "0"}, {"exchange_rate", "-10"}, {"exchange_rate", "1e3"}, {"exchange_rate", "1.0000001"}, {"exchange_rate", "1000001"},
	} {
		in := gatewayDraft()
		in.Values = map[string]string{tc.field: tc.value}
		before, _ := h.repo.Snapshot(ctx)
		_, err := h.s.SaveGatewaySettings(ctx, h.admin, in, h.key())
		var fieldErr *backoffice.GatewayFieldError
		if !errors.As(err, &fieldErr) || fieldErr.Fields["values."+tc.field] == "" {
			t.Fatalf("%+v: %v", tc, err)
		}
		after, _ := h.repo.Snapshot(ctx)
		if !reflect.DeepEqual(before.Commerce, after.Commerce) {
			t.Fatal("invalid rate mutated state")
		}
	}
}
