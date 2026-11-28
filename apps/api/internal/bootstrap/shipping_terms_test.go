package bootstrap_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

func TestShippingTermsPostgresTwoShipmentsRestartAndSellerScope(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := httpCapacitySeed(t)
	for i := range seed.Access.Roles {
		if seed.Access.Roles[i].ID == "store-shipping-test" {
			seed.Access.Roles[i].Permissions = append(seed.Access.Roles[i].Permissions, "orders.read")
		}
	}
	seed.Offers = append(seed.Offers, catalog.Offer{ID: "offer-two", ProductID: "product", VariantID: "variant", StoreID: "s2", Price: 1200000, Stock: 10, Version: 1})
	seed.Commerce.Settings.Shipping[0].CollectionMode = "collect"
	seed.Commerce.Settings.Shipping[0].Fee = 0
	second := seed.Commerce.Settings.Shipping[0]
	second.ID = "range"
	second.Name = "رایگان بدون ساعت"
	second.CollectionMode = "prepaid"
	second.DeliveryMode = "range"
	second.Schedule = nil
	second.MinDays = 1
	second.MaxDays = 3
	seed.Commerce.Settings.Shipping = append(seed.Commerce.Settings.Shipping, second)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	clock := capacityHTTPClock{now: time.Date(2026, 11, 21, 8, 0, 0, 0, shipping.Tehran())}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo, clock)
	admin := login(t, h, "09000000001")
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	var cart backoffice.CartView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/cart", nil, "", 200), &cart)
	for _, offer := range []string{"offer", "offer-two"} {
		json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: offer, Quantity: 1, Version: cart.Version}, offer, 200), &cart)
	}
	var address commerce.Address
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", backoffice.AddressInput{HouseNumber: "12", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Recipient: "گیرنده آزمون", Phone: "09000000003", Province: "تهران", City: "تهران", Street: "نشانی ساختگی", PostalCode: "1234567890"}, "address", 200), &address)
	var options []backoffice.StoreShippingChoices
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/checkout/shipping-options?address_id="+address.ID, nil, "", 200), &options)
	if len(options) != 2 {
		t.Fatal(options)
	}
	in := backoffice.QuoteInput{ContractVersion: 2, CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post", "s2": "range"}, PaymentMethodID: "card-transfer", WalletToman: "0", ShippingSelections: map[string]shipping.DeliverySelection{"s1": {Date: "2026-11-21", WindowID: "test-window"}}}
	var q backoffice.QuoteView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/quotes", in, "quote", 200), &q)
	if q.ShippingToman != "0" || len(q.Shipments) != 2 {
		t.Fatal(q)
	}
	var p backoffice.PurchaseView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, "accept", 200), &p)
	var d backoffice.DepositView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "card-transfer", PurchaseID: p.ID, DestinationID: "test-card", AmountToman: q.TotalToman, Date: clock.Now().Format("2006-01-02"), Reference: "SHIP123-FUND"}, "fund", 200), &d)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/deposits/review", backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedAmountToman: d.AmountToman, VerifiedReference: d.Reference, VerifiedDate: d.Date}, "approve", 200)
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "confirm", 200), &p)
	if p.Status != "confirmed" {
		t.Fatal(p)
	}
	before, _ := repo.Snapshot(ctx)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].Name = "نام جدید"
		w.Commerce.Settings.Shipping[0].MinDays = 20
		w.Stores[0].LeadDays = 9
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	restarted, err := reviewpostgres.New(ctx, db, seed)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := restarted.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Commerce.Quotes, after.Commerce.Quotes) {
		t.Fatal("history overwritten")
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted, clock)
	buyer = loginAt(t, h, "storefront", "09000000003")["token"].(string)
	var restored backoffice.PurchaseView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/orders/"+p.Number, nil, "", 200), &restored)
	for i := range p.Quote.Shipments {
		p.Quote.Shipments[i].StoreID = backoffice.PublicReference("s", p.Quote.Shipments[i].StoreID)
	}
	if !reflect.DeepEqual(p.Quote.Shipments, restored.Quote.Shipments) {
		t.Fatalf("promise drift: before=%+v after=%+v", p.Quote.Shipments, restored.Quote.Shipments)
	}
	seller := sellerLogin(t, h, "09120000002")
	var scoped []backoffice.SellerCheckoutOrder
	json.Unmarshal(commerceHTTP(t, h, seller, "", "GET", "/seller/checkout/orders", nil, "", 200), &scoped)
	if len(scoped) != 1 || scoped[0].StoreID != "s1" || scoped[0].CollectionMode != "collect" || scoped[0].Delivery == nil || scoped[0].Delivery.Slot == nil || scoped[0].ShippingFeeToman != "" {
		t.Fatal(scoped)
	}
	commerceHTTP(t, h, seller, "", "GET", "/seller/checkout/orders/"+p.ID+":s2", nil, "", 404)
}
