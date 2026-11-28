package bootstrap_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
)

func TestGatewayOAuthPostgresRestartEditWithoutMerchantAndPrivateResponses(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = commerceSeed(); return nil }); err != nil {
		t.Fatal(err)
	}
	cfg := gatewaySettingsConfig()
	h := bootstrap.NewHTTPHandlerWithRepository(cfg, repo)
	admin := login(t, h, "09000000001")
	secret := "OAuth-MiXeD-fixture-" + strings.Repeat("Qz9+/_=", 70)
	in := gatewaySettingsFixture()
	in.Credentials = map[string]string{"client_id": "۱۲۳۴", "client_secret": secret}
	in.Values = map[string]string{"currency": "IRR", "exchange_rate": "۱۰.۲۵"}
	raw := commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/gateways", in, "oauth-only", 200)
	var saved backoffice.GatewaySettingsView
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.ConfiguredCredentials) != 2 || strings.Contains(string(raw), secret) {
		t.Fatal("OAuth save leaked/lost data")
	}
	var payload []byte
	if err := db.QueryRow(ctx, "SELECT payload FROM review_workspace WHERE singleton").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), secret) {
		t.Fatal("secret stored in plaintext")
	}
	reopened, err := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(cfg, reopened)
	a2 := login(t, h2, "09000000001")
	var directory backoffice.GatewaySettingsDirectory
	read := commerceHTTP(t, h2, a2, "", "GET", "/backoffice/commerce/gateways", nil, "", 200)
	json.Unmarshal(read, &directory)
	if len(directory.Items) != 1 || len(directory.Items[0].ConfiguredCredentials) != 2 || strings.Contains(string(read), secret) {
		t.Fatal("restart lost or exposed OAuth settings")
	}
	if directory.Items[0].Values["currency"] != "IRR" || directory.Items[0].Values["exchange_rate"] != "10.25" {
		t.Fatal("restart lost gateway currency/rate")
	}
	edit := directory.Items[0].GatewaySettingsInput
	edit.Name = "اصلاح زنده تنظیمات بدون build"
	edit.Credentials = map[string]string{"client_secret": ""}
	commerceHTTP(t, h2, a2, "", "POST", "/backoffice/commerce/gateways", edit, "oauth-preserve", 200)
	state, _ := reopened.Snapshot(ctx)
	if len(state.Commerce.GatewaySettings[0].CredentialNames) != 2 || state.Commerce.GatewaySettings[0].Version != 2 {
		t.Fatal("blank edit removed credentials")
	}
	for _, path := range []string{"/storefront/commerce/config", "/storefront/home"} {
		public := commerceHTTP(t, h2, "", "", "GET", path, nil, "", 200)
		if strings.Contains(string(public), secret) || strings.Contains(string(public), "encrypted_credentials") {
			t.Fatal("secret leaked to storefront")
		}
	}
	bad := saved.GatewaySettingsInput
	bad.Version = 2
	bad.Credentials = map[string]string{"password": "ONE-TIME-CODE"}
	commerceHTTP(t, h2, a2, "", "POST", "/backoffice/commerce/gateways", bad, "no-otp-storage", 400)
}
