package backoffice_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const gatewayTestMerchant = "01234567-89ab-4cde-8fab-0123456789ab"

func gatewayDraft() backoffice.GatewaySettingsInput {
	return backoffice.GatewaySettingsInput{Provider: "zarinpal", Name: "زرین‌پال آزمایشی", MinToman: "1000", MaxToman: "1000000", AllCategories: true, MaxAutomaticRetries: 3, AllowWalletTopup: true, Credentials: map[string]string{"merchant_id": gatewayTestMerchant}}
}
func configureGatewayKey(t *testing.T, h *commerceHarness) {
	t.Helper()
	if err := h.s.ConfigureGatewaySettingsKey(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{37}, 32))); err != nil {
		t.Fatal(err)
	}
}
func TestGatewaySettingsEncryptedPrivateVersionedAndDoNotActivatePayments(t *testing.T) {
	h := newCommerceHarness(t)
	configureGatewayKey(t, h)
	ctx := context.Background()
	before, _ := h.repo.Snapshot(ctx)
	input := gatewayDraft()
	key := h.key()
	raw, err := h.s.SaveGatewaySettings(ctx, h.admin, input, key)
	saved := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if saved.ID == "" || saved.Version != 1 || saved.ExecutionReady || saved.Active || saved.MaxAutomaticRetries != 3 || !saved.AllowWalletTopup || len(saved.ConfiguredCredentials) != 1 || saved.Values["environment"] != "sandbox" {
		t.Fatal(saved)
	}
	if strings.Contains(string(raw), gatewayTestMerchant) || len(saved.Credentials) != 0 {
		t.Fatal("credential returned")
	}
	retry, e := h.s.SaveGatewaySettings(ctx, h.admin, input, key)
	if e != nil || string(retry) != string(raw) {
		t.Fatal("create replay", e)
	}
	state, _ := h.repo.Snapshot(ctx)
	if len(state.Commerce.GatewaySettings) != 1 {
		t.Fatal("duplicate draft")
	}
	stored := state.Commerce.GatewaySettings[0]
	encoded, _ := json.Marshal(state)
	if strings.Contains(string(encoded), gatewayTestMerchant) || stored.EncryptedCredentials == "" {
		t.Fatal("plaintext credential persisted")
	}
	for _, pair := range []struct {
		name          string
		before, after any
	}{{"methods", before.Commerce.Settings.Methods, state.Commerce.Settings.Methods}, {"entries", before.Commerce.Entries, state.Commerce.Entries}, {"deposits", before.Commerce.Deposits, state.Commerce.Deposits}, {"purchases", before.Commerce.Purchases, state.Commerce.Purchases}, {"orders", before.Orders, state.Orders}} {
		if !reflect.DeepEqual(pair.before, pair.after) {
			t.Fatal("modified " + pair.name)
		}
	}
	dir, e := h.s.ReadGatewaySettings(ctx, h.admin)
	if e != nil || !dir.CanWrite || !dir.CredentialsWritable || len(dir.Items) != 1 {
		t.Fatal(dir, e)
	}
	public, e := h.s.PublicCommerceConfig(ctx)
	if e != nil {
		t.Fatal(e)
	}
	pub, _ := json.Marshal(public)
	if strings.Contains(string(pub), "zarinpal") || strings.Contains(string(pub), "gateway_settings") || strings.Contains(string(pub), gatewayTestMerchant) {
		t.Fatal("private gateway leaked")
	}
	wallet := h.wallet(t, h.buyer)
	if len(wallet.Methods) != 2 {
		t.Fatal("draft became a funding method", wallet.Methods)
	}
	edit := saved.GatewaySettingsInput
	edit.MaxAutomaticRetries = 0
	edit.AllowWalletTopup = false
	edit.Name = "ویرایش پیش‌نویس"
	edit.Credentials = map[string]string{"merchant_id": ""}
	raw, e = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	updated := decodeCommerce[backoffice.GatewaySettingsView](t, raw, e)
	if updated.Version != 2 || updated.MaxAutomaticRetries != 0 || updated.AllowWalletTopup || len(updated.ConfiguredCredentials) != 1 {
		t.Fatal("blank overwrote credential", updated)
	}
	if _, e = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key()); e != backoffice.ErrConflict {
		t.Fatal("stale version", e)
	}
	edit = updated.GatewaySettingsInput
	edit.ClearCredentials = []string{"merchant_id"}
	raw, e = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	cleared := decodeCommerce[backoffice.GatewaySettingsView](t, raw, e)
	if len(cleared.ConfiguredCredentials) != 0 {
		t.Fatal("explicit clear failed")
	}
}

func TestGatewaySettingsValidationAndAuthorization(t *testing.T) {
	h := newCommerceHarness(t)
	configureGatewayKey(t, h)
	ctx := context.Background()
	for _, tc := range []struct {
		name   string
		change func(*backoffice.GatewaySettingsInput)
	}{
		{"negative retries", func(v *backoffice.GatewaySettingsInput) { v.MaxAutomaticRetries = -1 }},
		{"too many retries", func(v *backoffice.GatewaySettingsInput) { v.MaxAutomaticRetries = 11 }},
		{"unsupported provider", func(v *backoffice.GatewaySettingsInput) { v.Provider = "other" }},
		{"bad credential", func(v *backoffice.GatewaySettingsInput) {
			v.Credentials = map[string]string{"merchant_id": "مدیر سایت"}
		}},
		{"unrelated credential", func(v *backoffice.GatewaySettingsInput) {
			v.Credentials = map[string]string{"password": "never-persist-this"}
		}},
		{"unrelated public field", func(v *backoffice.GatewaySettingsInput) { v.Values = map[string]string{"apikey": "never-persist-this"} }},
		{"foreign service URL", func(v *backoffice.GatewaySettingsInput) {
			v.Values = map[string]string{"base_url": "https://foreign.invalid/"}
		}},
		{"wrong environment URL", func(v *backoffice.GatewaySettingsInput) { v.Values = map[string]string{"environment": "production"} }},
		{"unknown currency", func(v *backoffice.GatewaySettingsInput) { v.Values = map[string]string{"currency": "USD"} }},
		{"invalid category", func(v *backoffice.GatewaySettingsInput) { v.AllowedCategories = []string{"missing"} }},
		{"wrong max", func(v *backoffice.GatewaySettingsInput) { v.MaxToman = "1" }},
		{"invalid OAuth client ID", func(v *backoffice.GatewaySettingsInput) {
			v.Credentials = map[string]string{"client_id": "not-an-integer"}
		}},
		{"oversized client secret", func(v *backoffice.GatewaySettingsInput) {
			v.Credentials = map[string]string{"client_secret": strings.Repeat("A", 4097)}
		}},
		{"OAuth credential in plain values", func(v *backoffice.GatewaySettingsInput) {
			v.Values = map[string]string{"client_secret": "never-store-in-values"}
		}},
		{"foreign OAuth API URL", func(v *backoffice.GatewaySettingsInput) {
			v.Values = map[string]string{"api_base_url": "https://foreign.invalid/api"}
		}},
		{"unknown OTP channel", func(v *backoffice.GatewaySettingsInput) { v.Values = map[string]string{"channel": "other"} }},
		{"callback cannot be overridden by administrator", func(v *backoffice.GatewaySettingsInput) {
			v.Values = map[string]string{"callback_url": "https://example.test/return"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := gatewayDraft()
			tc.change(&in)
			before, _ := h.repo.Snapshot(ctx)
			_, err := h.s.SaveGatewaySettings(ctx, h.admin, in, h.key())
			var fieldErr *backoffice.GatewayFieldError
			if !errors.As(err, &fieldErr) {
				t.Fatal(err)
			}
			after, _ := h.repo.Snapshot(ctx)
			if !reflect.DeepEqual(before.Commerce, after.Commerce) {
				t.Fatal("invalid draft changed state")
			}
		})
	}
	for _, actor := range []identity.User{h.ops, h.finance, h.buyer, func() identity.User { v := h.admin; v.Surface = identity.SurfaceSeller; return v }()} {
		if _, err := h.s.SaveGatewaySettings(ctx, actor, gatewayDraft(), h.key()); err != backoffice.ErrForbidden {
			t.Fatal("unauthorized write", actor.Surface, err)
		}
	}
	read, err := h.s.ReadGatewaySettings(ctx, h.finance)
	if err != nil || read.CanWrite || read.CredentialsWritable {
		t.Fatal("read-only grant", read, err)
	}
	if _, err = h.s.ReadGatewaySettings(ctx, h.buyer); err != backoffice.ErrForbidden {
		t.Fatal("customer read", err)
	}
	valid := gatewayDraft()
	valid.MaxAutomaticRetries = 10
	if _, err = h.s.SaveGatewaySettings(ctx, h.admin, valid, h.key()); err != nil {
		t.Fatal("maximum retry value", err)
	}
}

func TestGatewaySettingsKeyRequiredOnlyForCredentialAndWrongKeyCannotOverwrite(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	input := gatewayDraft()
	if _, err := h.s.SaveGatewaySettings(ctx, h.admin, input, h.key()); !errors.Is(err, backoffice.ErrGatewayKeyUnavailable) {
		t.Fatal(err)
	}
	input.Credentials = nil
	raw, err := h.s.SaveGatewaySettings(ctx, h.admin, input, h.key())
	draft := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	configureGatewayKey(t, h)
	edit := draft.GatewaySettingsInput
	edit.Credentials = map[string]string{"merchant_id": gatewayTestMerchant}
	raw, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	saved := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	before, _ := h.repo.Snapshot(ctx)
	if err = h.s.ConfigureGatewaySettingsKey(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{99}, 32))); err != nil {
		t.Fatal(err)
	}
	if _, err = h.s.SaveGatewaySettings(ctx, h.admin, saved.GatewaySettingsInput, h.key()); !errors.Is(err, backoffice.ErrGatewayKeyUnavailable) {
		t.Fatal("wrong key", err)
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Commerce, after.Commerce) {
		t.Fatal("wrong key changed state")
	}
	if err = h.s.ConfigureGatewaySettingsKey("bad-key"); !errors.Is(err, backoffice.ErrGatewayKeyUnavailable) {
		t.Fatal("invalid key accepted")
	}
}
