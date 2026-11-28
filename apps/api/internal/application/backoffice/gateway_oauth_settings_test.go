package backoffice_test

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func decryptGatewayFixture(t *testing.T, g commerce.GatewaySettings) map[string]string {
	t.Helper()
	block, err := aes.NewCipher(bytes.Repeat([]byte{37}, 32))
	if err != nil {
		t.Fatal(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(g.EncryptedCredentials)
	if err != nil {
		t.Fatal(err)
	}
	n := aead.NonceSize()
	if len(raw) < n {
		t.Fatal("missing encrypted value")
	}
	plain, err := aead.Open(nil, raw[:n], raw[n:], []byte("cherriva:gateway-settings:v1:"+g.ID+":"+g.Provider))
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]string
	if err = json.Unmarshal(plain, &values); err != nil {
		t.Fatal(err)
	}
	return values
}

func TestGatewayOAuthCredentialsWithoutMerchantPreserveOpaqueSecretAndIndependentFields(t *testing.T) {
	h := newCommerceHarness(t)
	configureGatewayKey(t, h)
	ctx := context.Background()
	secret := "  MiXeD-OAuth-" + strings.Repeat("Az9+/_=", 80) + "  "
	in := gatewayDraft()
	in.Credentials = map[string]string{"client_id": "۱۲۳۴", "client_secret": secret, "username": "merchant-fixture@example.test"}
	raw, err := h.s.SaveGatewaySettings(ctx, h.admin, in, h.key())
	saved := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if len(saved.ConfiguredCredentials) != 3 || strings.Contains(string(raw), secret) || strings.Contains(string(raw), "merchant-fixture@example.test") {
		t.Fatal("OAuth values exposed")
	}
	if _, ok := saved.Values["client_secret"]; ok {
		t.Fatal("credential stored as a normal value")
	}
	state, _ := h.repo.Snapshot(ctx)
	values := decryptGatewayFixture(t, state.Commerce.GatewaySettings[0])
	if values["client_id"] != "1234" || values["client_secret"] != secret || values["username"] != "merchant-fixture@example.test" || values["merchant_id"] != "" {
		t.Fatal("OAuth data was transformed or merchant required")
	}
	for _, p := range backoffice.GatewayProviders() {
		for _, f := range p.Fields {
			if f.Key == "merchant_id" && f.RequiredForActivation {
				t.Fatal("merchant ID still universally required")
			}
			if f.Key == "password" || f.Key == "description" || f.Key == "callback_url" {
				t.Fatal("operation-specific field exposed as a setting", f.Key)
			}
		}
	}
	edit := saved.GatewaySettingsInput
	edit.Name = "ویرایش بدون انتشار"
	edit.Credentials = map[string]string{"client_secret": "", "client_id": ""}
	raw, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	saved = decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	state, _ = h.repo.Snapshot(ctx)
	if got := decryptGatewayFixture(t, state.Commerce.GatewaySettings[0]); !reflect.DeepEqual(values, got) {
		t.Fatal("blank edit replaced stored fields")
	}
	edit = saved.GatewaySettingsInput
	edit.Credentials = map[string]string{"client_id": "5678"}
	raw, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	saved = decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	state, _ = h.repo.Snapshot(ctx)
	values = decryptGatewayFixture(t, state.Commerce.GatewaySettings[0])
	if values["client_id"] != "5678" || values["client_secret"] != secret {
		t.Fatal("independent update altered secret")
	}
	edit = saved.GatewaySettingsInput
	edit.ClearCredentials = []string{"client_secret"}
	_, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	if err != nil {
		t.Fatal(err)
	}
	state, _ = h.repo.Snapshot(ctx)
	values = decryptGatewayFixture(t, state.Commerce.GatewaySettings[0])
	if values["client_secret"] != "" || values["client_id"] != "5678" || values["username"] == "" {
		t.Fatal("independent clear failed")
	}
}

func TestGatewayLegacyDraftKeepsCredentialAndDropsStaticDescriptionFromViewsAndEdits(t *testing.T) {
	h := newCommerceHarness(t)
	configureGatewayKey(t, h)
	ctx := context.Background()
	raw, err := h.s.SaveGatewaySettings(ctx, h.admin, gatewayDraft(), h.key())
	saved := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if err = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.GatewaySettings[0].Values["description"] = "شرح ثابت قدیمی"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	dir, err := h.s.ReadGatewaySettings(ctx, h.admin)
	if err != nil {
		t.Fatal(err)
	}
	view := dir.Items[0]
	if _, exists := view.Values["description"]; exists {
		t.Fatal("legacy static description exposed")
	}
	if view.Values["api_base_url"] != "https://next.zarinpal.com/api" || view.Values["channel"] != "ussd" {
		t.Fatal("new defaults missing on existing draft")
	}
	edit := view.GatewaySettingsInput
	edit.Values["description"] = "ارسال از فرم قدیمی"
	edit.Credentials = map[string]string{"client_id": "71", "client_secret": "New-Case-Sensitive-Value"}
	if _, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key()); err != nil {
		t.Fatal(err)
	}
	state, _ := h.repo.Snapshot(ctx)
	g := state.Commerce.GatewaySettings[0]
	values := decryptGatewayFixture(t, g)
	if g.ID != saved.ID || g.Version != 2 || values["merchant_id"] != gatewayTestMerchant || values["client_secret"] != "New-Case-Sensitive-Value" {
		t.Fatal("legacy data lost")
	}
	if _, exists := g.Values["description"]; exists {
		t.Fatal("static description persisted after edit")
	}
}

func TestGatewayTransactionDescriptionDependsOnOperation(t *testing.T) {
	charge, err := backoffice.GatewayTransactionDescription("wallet-topup", "")
	if err != nil || charge != "شارژ کیف پول چریوا" {
		t.Fatal(charge, err)
	}
	order, err := backoffice.GatewayTransactionDescription("order-payment", "1014")
	if err != nil || order != "پرداخت سفارش شماره 1014" || charge == order {
		t.Fatal(order, err)
	}
	for _, args := range [][2]string{{"order-payment", ""}, {"order-payment", "متن دلخواه"}, {"wallet-topup", "1014"}, {"unknown", "1014"}} {
		if _, err = backoffice.GatewayTransactionDescription(args[0], args[1]); err != backoffice.ErrInvalid {
			t.Fatal("invalid operation accepted", args, err)
		}
	}
}
