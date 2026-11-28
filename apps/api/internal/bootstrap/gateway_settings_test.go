package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

const gatewayMerchantFixture = "01234567-89ab-4cde-8fab-0123456789ab"

func gatewaySettingsFixture() backoffice.GatewaySettingsInput {
	return backoffice.GatewaySettingsInput{Provider: "zarinpal", Name: "درگاه آزمون مستقل", MinToman: "1000", MaxAutomaticRetries: 4, AllowWalletTopup: true, AllCategories: true, Credentials: map[string]string{"merchant_id": gatewayMerchantFixture}}
}
func gatewaySettingsConfig() config.Config {
	return config.Config{Environment: "test", PaymentGatewayKey: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{31}, 32))}
}

func TestGatewaySettingsHTTPPermissionsFieldErrorsAndPublicIsolation(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(gatewaySettingsConfig(), repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	for _, token := range []string{"", seller} {
		commerceHTTP(t, h, token, "", "GET", "/backoffice/commerce/gateways", nil, "", 401)
	}
	commerceHTTP(t, h, viewer, "", "POST", "/backoffice/commerce/gateways", gatewaySettingsFixture(), "denied-gateway", 403)
	bad := gatewaySettingsFixture()
	bad.MaxAutomaticRetries = 11
	result := commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", bad, "bad-retries", 400)
	if !strings.Contains(string(result), "max_automatic_retries") {
		t.Fatal("missing field error", string(result))
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", map[string]any{"provider": "zarinpal", "unknown": "value"}, "unknown-field", 400)
	raw := commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", gatewaySettingsFixture(), "save-gateway", 200)
	var saved backoffice.GatewaySettingsView
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Version != 1 || saved.ID == "" || strings.Contains(string(raw), gatewayMerchantFixture) || len(saved.ConfiguredCredentials) != 1 {
		t.Fatal("response leak", string(raw))
	}
	if replay := commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", gatewaySettingsFixture(), "save-gateway", 200); string(replay) != string(raw) {
		t.Fatal("replay changed")
	}
	for _, path := range []string{"/backoffice/commerce/gateways", "/backoffice/commerce/settings", "/backoffice/commerce"} {
		body := commerceHTTP(t, h, admin, "", "GET", path, nil, "", 200)
		if strings.Contains(string(body), gatewayMerchantFixture) || strings.Contains(string(body), "encrypted_credentials") {
			t.Fatal("admin read leaked", path)
		}
	}
	public := commerceHTTP(t, h, "", "", "GET", "/storefront/commerce/config", nil, "", 200)
	if strings.Contains(string(public), "zarinpal") || strings.Contains(string(public), gatewayMerchantFixture) {
		t.Fatal("public draft leak")
	}
	stale := saved.GatewaySettingsInput
	stale.Version = 0
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", stale, "stale", 409)
	bad = saved.GatewaySettingsInput
	bad.Active = true
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", bad, "activate-settings", 200), &saved)
	hNoKey := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	adminNoKey := login(t, hNoKey, "09000000001")
	commerceHTTP(t, hNoKey, adminNoKey, "", "POST", "/backoffice/commerce/gateways", saved.GatewaySettingsInput, "missing-key", 503)
}

func TestGatewaySettingsPostgresPersistenceRollbackReplayAndUnchangedFinancialState(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = commerceSeed(); return nil }); err != nil {
		t.Fatal(err)
	}
	before, _ := repo.Snapshot(ctx)
	cfg := gatewaySettingsConfig()
	h := bootstrap.NewHTTPHandlerWithRepository(cfg, repo)
	admin := login(t, h, "09000000001")
	raw := commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", gatewaySettingsFixture(), "pg-gateway", 200)
	var saved backoffice.GatewaySettingsView
	json.Unmarshal(raw, &saved)
	var payload []byte
	if err := db.QueryRow(ctx, "SELECT payload FROM review_workspace WHERE singleton").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), gatewayMerchantFixture) || !strings.Contains(string(payload), "encrypted_credentials") {
		t.Fatal("credential encryption missing")
	}
	reopened, err := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(cfg, reopened)
	a2 := login(t, h2, "09000000001")
	read := commerceHTTP(t, h2, a2, "", "GET", "/backoffice/commerce/gateways", nil, "", 200)
	if !strings.Contains(string(read), saved.ID) || strings.Contains(string(read), gatewayMerchantFixture) {
		t.Fatal("reopen failed/leaked")
	}
	commerceHTTP(t, h2, a2, "", "POST", "/backoffice/commerce/gateways", gatewaySettingsFixture(), "pg-gateway", 200)
	edit := saved.GatewaySettingsInput
	edit.AllowWalletTopup = false
	edit.MaxAutomaticRetries = 0
	_, err = db.Exec(ctx, `CREATE FUNCTION fail_gateway_settings() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'isolated gateway settings persistence failure'; END $$; CREATE TRIGGER fail_gateway_settings BEFORE UPDATE ON review_workspace FOR EACH ROW EXECUTE FUNCTION fail_gateway_settings()`)
	if err != nil {
		t.Fatal(err)
	}
	commerceHTTP(t, h2, a2, "", "POST", "/backoffice/commerce/gateways", edit, "edit-gateway", 503)
	_, err = db.Exec(ctx, `DROP TRIGGER fail_gateway_settings ON review_workspace; DROP FUNCTION fail_gateway_settings()`)
	if err != nil {
		t.Fatal(err)
	}
	afterFailure, _ := repo.Snapshot(ctx)
	if afterFailure.Commerce.GatewaySettings[0].Version != 1 || !afterFailure.Commerce.GatewaySettings[0].AllowWalletTopup {
		t.Fatal("failed save committed")
	}
	commerceHTTP(t, h2, a2, "", "POST", "/backoffice/commerce/gateways", edit, "edit-gateway", 200)
	after, _ := reopened.Snapshot(ctx)
	if len(after.Commerce.GatewaySettings) != 1 || after.Commerce.GatewaySettings[0].Version != 2 || after.Commerce.GatewaySettings[0].AllowWalletTopup || after.Commerce.GatewaySettings[0].MaxAutomaticRetries != 0 {
		t.Fatal("saved policy not persisted")
	}
	for name, pair := range map[string][2]any{"methods": {before.Commerce.Settings.Methods, after.Commerce.Settings.Methods}, "bank": {before.Commerce.Settings.Destination, after.Commerce.Settings.Destination}, "entries": {before.Commerce.Entries, after.Commerce.Entries}, "deposits": {before.Commerce.Deposits, after.Commerce.Deposits}, "orders": {before.Orders, after.Orders}, "purchases": {before.Commerce.Purchases, after.Commerce.Purchases}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("changed " + name)
		}
	}
}
