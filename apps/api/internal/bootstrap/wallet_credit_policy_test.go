package bootstrap_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func TestAssignableWalletCreditPermissionAndRevocation(t *testing.T) {
	seed := commerceSeed()
	var operator, target string
	for _, u := range seed.Users {
		if u.Phone == "09000000002" {
			operator = u.ID
		}
		if u.Phone == "09000000003" {
			target = u.ID
		}
	}
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "credit-only", Name: "شارژ مستقل آزمون", Scope: identity.ScopeGlobal, Permissions: []string{backoffice.CommerceWalletCredit}, Version: 1})
	seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: operator, RoleID: "credit-only"})
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	token := login(t, h, "09000000002")
	in := backoffice.AdminCreditInput{UserID: target, AmountToman: "10", Title: "مجوز آزمون", Description: "شارژ با نقش قابل انتساب"}
	commerceHTTP(t, h, token, "", "POST", "/backoffice/commerce/wallet/credits", in, "custom-credit", 200)
	var workspace backoffice.Workspace
	json.Unmarshal(commerceHTTP(t, h, token, "", "GET", "/backoffice/workspace", nil, "", 200), &workspace)
	if len(workspace.Users) == 0 || workspace.Commerce != nil {
		t.Fatal("customer list missing or financial state leaked")
	}
	if e := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i, g := range w.Access.Grants {
			if g.UserID == operator && g.RoleID == "credit-only" {
				w.Access.Grants = append(w.Access.Grants[:i], w.Access.Grants[i+1:]...)
				break
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	commerceHTTP(t, h, token, "", "POST", "/backoffice/commerce/wallet/credits", in, "revoked-credit", 403)
}

func TestAdminWalletCreditHTTPValidationAndOperationRecovery(t *testing.T) {
	seed := commerceSeed()
	var buyerID string
	for _, user := range seed.Users {
		if user.Phone == "09000000003" {
			buyerID = user.ID
		}
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(seed))
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	if buyerID == "" {
		t.Fatal("missing buyer ID")
	}
	path := "/backoffice/commerce/wallet/credits"
	in := backoffice.AdminCreditInput{UserID: buyerID, AmountToman: "123.4", Title: "شارژ آزمون", Description: "دلیل ساختگی برای بررسی"}
	commerceHTTP(t, h, "", "", "POST", path, in, "no-session", 401)
	commerceHTTP(t, h, buyer, "", "POST", path, in, "wrong-surface", 401)
	commerceHTTP(t, h, viewer, "", "POST", path, in, "no-permission", 403)
	commerceHTTP(t, h, admin, "", "POST", path, map[string]any{"user_id": buyerID, "amount_toman": "10", "title": "آزمون", "description": "ساختگی", "extra": true}, "unknown-field", 400)
	first := commerceHTTP(t, h, admin, "", "POST", path, in, "credit", 200)
	if string(first) != string(commerceHTTP(t, h, admin, "", "POST", path, in, "credit", 200)) {
		t.Fatal("idempotent result changed")
	}
	var got backoffice.AdminCreditWallet
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", path+"?user_id="+buyerID, nil, "credit", 200), &got)
	if got.PreviousOperation == nil || got.PreviousOperation.AmountToman != "123.4" || got.PreviousOperation.Withdrawable || got.BalanceToman != "123.4" {
		t.Fatal(got)
	}
	in.AmountToman = "200"
	commerceHTTP(t, h, admin, "", "POST", path, in, "credit", 409)
	commerceHTTP(t, h, viewer, "", "GET", path+"?user_id="+buyerID, nil, "", 403)
	commerceHTTP(t, h, admin, "", "GET", path+"?user_id=missing", nil, "", 404)
	r := httptest.NewRequest("GET", path+"?user_id="+buyerID, nil)
	r.Header.Set("Authorization", "Bearer "+admin)
	r.Header.Set("Idempotency-Key", strings.Repeat("x", 129))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/refunds", backoffice.RefundInput{AmountToman: "1", Instructions: "مقصد آزمون"}, "restricted-refund", 409)
}

func TestWalletCreditPostgresLegacyRestartAndImmutableSources(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := commerceSeed()
	var buyerID string
	for _, u := range seed.Users {
		if u.Phone == "09000000003" {
			buyerID = u.ID
		}
	}
	if buyerID == "" {
		t.Fatal("fixture buyer")
	}
	legacy := commerce.WalletEntry{ID: "legacy-policy", UserID: buyerID, Source: "legacy-policy", Amount: 2000000, Kind: "deposit-credit", At: time.Now().UTC().Add(-24 * time.Hour)}
	seed.Commerce.Entries = []commerce.WalletEntry{legacy}
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	in := backoffice.AdminCreditInput{UserID: buyerID, AmountToman: "50000", Title: "هدیه آزمون SQL", Description: "منبع محدود ساختگی"}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/wallet/credits", in, "pg-credit", 200)
	before, e := repo.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if len(before.Commerce.Entries) != 2 || before.Commerce.Entries[0].ID != legacy.ID || before.Commerce.Entries[0].Amount != legacy.Amount {
		t.Fatal("legacy ledger rewritten")
	}
	again, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, again)
	admin = login(t, h, "09000000001")
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	var wallet backoffice.WalletView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/wallet", nil, "", 200), &wallet)
	if wallet.BalanceToman != "250000" || wallet.WithdrawableToman != "200000" || wallet.PurchaseOnlyToman != "50000" {
		t.Fatal(wallet)
	}
	// Exercise split cancellation sources through the relational ledger/FK guards.
	var cart backoffice.CartView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/cart", nil, "", 200), &cart)
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: "offer", Quantity: 1, Version: cart.Version}, "policy-cart", 200), &cart)
	var address commerce.Address
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", backoffice.AddressInput{HouseNumber: "12", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Recipient: "گیرنده آزمون", Phone: "09000000003", Province: "تهران", City: "تهران", Street: "نشانی مستقل آزمون SQL", PostalCode: "1234567890"}, "policy-address", 200), &address)
	var quote backoffice.QuoteView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/quotes", backoffice.QuoteInput{ContractVersion: 2, WalletToman: "110000", CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post"}}, "policy-quote", 200), &quote)
	var purchase backoffice.PurchaseView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", backoffice.AcceptQuoteInput{QuoteID: quote.ID, AcceptedTotalToman: quote.TotalToman, TermsDigest: quote.TermsDigest}, "policy-accept", 200), &purchase)
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: purchase.ID, Version: purchase.Version}, "policy-confirm", 200), &purchase)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/orders/cancel", backoffice.PurchaseAdminInput{ID: purchase.ID, Version: purchase.Version, Reason: "لغو مستقل برای آزمون محدودیت منابع"}, "policy-cancel", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/wallet/credits", in, "pg-credit", 200)
	if e = again.Update(ctx, func(w *backoffice.Workspace) error {
		flag := true
		w.Commerce.CreditLots[1].Withdrawable = &flag
		return nil
	}); e == nil {
		t.Fatal("restriction mutation persisted")
	}
	if e = again.Update(ctx, func(w *backoffice.Workspace) error { w.Commerce.CreditLots = w.Commerce.CreditLots[:1]; return nil }); e == nil {
		t.Fatal("source deletion persisted")
	}
	var refund backoffice.RefundView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/refunds", backoffice.RefundInput{AmountToman: "200000", Instructions: "خروج نقد آزمون"}, "pg-refund", 200), &refund)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/refunds/resolve", backoffice.RefundResolveInput{ID: refund.ID, Version: refund.Version, Complete: true, Reference: "PG-POLICY-OUT", Date: time.Now().Format("2006-01-02"), Reason: "خروج ساختگی"}, "pg-refund-confirm", 200)
	if e = again.Update(ctx, func(w *backoffice.Workspace) error { w.Commerce.CreditLots[0].Spent = 0; return nil }); e == nil {
		t.Fatal("spent amount reversed")
	}
	var count int
	if e = db.QueryRow(ctx, "SELECT count(*) FROM commerce_credit_lots").Scan(&count); e != nil || count != 4 {
		t.Fatal(count, e)
	}
	var updatable bool
	if e = db.QueryRow(ctx, "SELECT withdrawable FROM commerce_credit_lots WHERE user_id=$1 AND origin_lot_id IS NULL AND amount=500000", buyerID).Scan(&updatable); e != nil || updatable {
		t.Fatal("SQL restriction", e)
	}
	down, e := os.ReadFile(filepath.Join("..", "..", "..", "..", "db", "migrations", "000010-wallet-credit-policy.down.sql"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, string(down)); e == nil {
		t.Fatal("rollback discarded financial source policy")
	}
	if _, e = reviewpostgres.New(ctx, db, backoffice.Workspace{}); e != nil {
		t.Fatal("restart after guarded rollback", e)
	}
}
