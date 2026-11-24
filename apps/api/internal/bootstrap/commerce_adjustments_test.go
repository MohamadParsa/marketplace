package bootstrap_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func TestCouponCreditHTTPAccessIdempotencyAndPostgresRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	// This helper owns an isolated schema; only its test fixture is replaced.
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = commerceSeed(); return nil }); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	viewer := login(t, h, "09000000005")
	call := func(token, method, path string, input any, key string, status int) []byte {
		return commerceHTTP(t, h, token, "", method, path, input, key, status)
	}
	in := backoffice.CouponInput{Code: "PG-CREDIT", Name: "کوپن آزمون پایدار", Active: true, Kind: "percent", Value: "10", Target: "order", AllCategories: true, MinToman: "0", MaxToman: "0", CapToman: "0", GlobalUnlimited: true, UserUnlimited: true, PlatformBPS: 10000}
	call("", "GET", "/backoffice/commerce/coupons", nil, "", 401)
	call(viewer, "POST", "/backoffice/commerce/coupons", in, "no-save", 403)
	call(buyer, "GET", "/backoffice/commerce/coupons", nil, "", 401)
	var coupon backoffice.CouponView
	json.Unmarshal(call(admin, "POST", "/backoffice/commerce/coupons", in, "save", 200), &coupon)
	call(admin, "POST", "/backoffice/commerce/coupons", in, "save", 200)
	in.Value = "20"
	call(admin, "POST", "/backoffice/commerce/coupons", in, "save", 409)
	call(viewer, "POST", "/backoffice/commerce/coupons/activate", backoffice.CouponActivationInput{ID: coupon.ID, Version: 1, Active: false}, "activate-denied", 403)
	var deposit backoffice.DepositView
	json.Unmarshal(call(buyer, "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "card-transfer", DestinationID: "test-card", AmountToman: "200000", Date: time.Now().Format("2006-01-02"), Reference: "PG-ADJUST-001"}, "deposit", 200), &deposit)
	review := backoffice.DepositReviewInput{ID: deposit.ID, Version: 1, Approve: true, VerifiedAmountToman: deposit.AmountToman, VerifiedDate: deposit.Date, VerifiedReference: deposit.Reference}
	call(admin, "POST", "/backoffice/commerce/deposits/review", review, "review", 200)
	var cart backoffice.CartView
	json.Unmarshal(call(buyer, "GET", "/storefront/cart", nil, "", 200), &cart)
	json.Unmarshal(call(buyer, "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: "offer", Quantity: 1, Version: cart.Version}, "cart", 200), &cart)
	var address commerce.Address
	json.Unmarshal(call(buyer, "POST", "/storefront/account/addresses", backoffice.AddressInput{Recipient: "گیرنده آزمایشی", Phone: "09000000003", Province: "تهران", City: "تهران", Street: "نشانی آزمایشی مستقل", PostalCode: "1234567890"}, "address", 200), &address)
	var q backoffice.QuoteView
	json.Unmarshal(call(buyer, "POST", "/storefront/checkout/quotes", backoffice.QuoteInput{ContractVersion: 2, CouponCode: coupon.Code, WalletToman: "99000", CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post"}}, "quote", 200), &q)
	if q.TotalToman != "99000" || q.DiscountToman != "11000" || q.FundingToman != "0" {
		t.Fatal(q)
	}
	accept := backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}
	call(buyer, "POST", "/storefront/checkout/purchases", accept, "bad-digest", 409)
	accept.TermsDigest = q.TermsDigest
	var p backoffice.PurchaseView
	json.Unmarshal(call(buyer, "POST", "/storefront/checkout/purchases", accept, "accept", 200), &p)
	call(buyer, "POST", "/storefront/checkout/purchases", accept, "accept", 200)
	before, e := repo.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	again, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, e := again.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before.Commerce, after.Commerce) {
		t.Fatal("coupon/hold/allocation changed across repository reopen")
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, again)
	admin = login(t, h, "09000000001")
	buyer = loginAt(t, h, "storefront", "09000000003")["token"].(string)
	call(admin, "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "confirm", 200)
	call(admin, "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "confirm-retry", 200)
	var wallet backoffice.WalletView
	json.Unmarshal(call(buyer, "GET", "/storefront/wallet", nil, "", 200), &wallet)
	if wallet.BalanceToman != "101000" || wallet.HeldToman != "0" {
		t.Fatal(wallet)
	}
	final, _ := again.Snapshot(ctx)
	count := 0
	for _, e := range final.Commerce.Entries {
		if e.Kind == "order-payment" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("double debit")
	}
	// Standard workspace cannot expose coupon or allocation data.
	raw := call(admin, "GET", "/backoffice/workspace", nil, "", 200)
	var public map[string]any
	json.Unmarshal(raw, &public)
	if public["commerce"] != nil {
		t.Fatal("private commerce leaked")
	}
}
func TestCouponHTTPRejectsUnknownFieldsAndCustomerScope(t *testing.T) {
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(commerceSeed()))
	admin := login(t, h, "09000000001")
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/coupons", map[string]any{"user_id": "other"}, "invalid", 400)
}

func TestGroupMappingInvalidatesPublicCacheAndPersists(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	w, _ := repo.Snapshot(ctx)
	// Existing fixture has a shared warranty axis; map its first two real options.
	var a catalog.AttributeDefinition
	for _, attr := range w.Attributes {
		if attr.Code == "warranty" {
			a = attr
		}
	}
	if len(a.Options) < 2 {
		t.Fatal("fixture warranty options")
	}
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	admin := login(t, h, "09000000001")
	stamp, e := repo.CatalogStamp(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	a.Groups = []catalog.AttributeValueGroup{{ID: "six-month", Name: "گارانتی نمایشی آزمون", Active: true, Values: a.Options[:2]}}
	productCall(t, h, "POST", "/backoffice/commands", admin, backoffice.Command{Action: "attribute.groups.set", Target: a.ID, Version: a.Version, Attribute: &a, IdempotencyKey: "group-map"}, 204)
	next, e := repo.CatalogStamp(ctx, "")
	if e != nil || next.Dependencies <= stamp.Dependencies {
		t.Fatal("mapping cache dependency unchanged", e)
	}
	again, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	restored, _ := again.Snapshot(ctx)
	for _, v := range restored.Attributes {
		if v.ID == a.ID && len(v.Groups) != 1 {
			t.Fatal("group missing after reopen")
		}
	}
}
