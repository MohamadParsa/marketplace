package bootstrap_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func TestPlacedOrderPostgresBeforeFundingAndReplayAfterRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := commerceSeed()
	seed.Offers = append(seed.Offers, catalog.Offer{ID: "second-offer", ProductID: "product", VariantID: "variant", StoreID: "s2", Price: 1200000, Stock: 10, Version: 1})
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
	cart := readJSON[backoffice.CartView](t, commerceHTTP(t, h, buyer, "", "GET", "/storefront/cart", nil, "", 200))
	for _, id := range []string{"offer", "second-offer"} {
		cart = readJSON[backoffice.CartView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: id, Quantity: 1, Version: cart.Version}, "cart-"+id, 200))
	}
	address := readJSON[commerce.Address](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", backoffice.AddressInput{Recipient: "گیرنده ساختگی", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Phone: "09000000111", Province: "تهران", City: "تهران", Street: "نشانی ساختگی برای آزمون", PostalCode: "1234567890", HouseNumber: "12"}, "address", 200))
	q := readJSON[backoffice.QuoteView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/quotes", backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post", "s2": "post"}, PaymentMethodID: "local"}, "quote", 200))
	bad := backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: "1"}
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", bad, "bad-amount", 409)
	w, _ := repo.Snapshot(ctx)
	if len(w.Orders) != 0 || len(w.Commerce.Purchases) != 0 || w.Offers[0].Reserved != 0 {
		t.Fatal("failed placement left partial state")
	}
	in := backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}
	first := commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", in, "place-before-funding", 200)
	p := readJSON[backoffice.PurchaseView](t, first)
	w, _ = repo.Snapshot(ctx)
	if len(p.OrderIDs) != 2 || p.OrderStatus != "placed" || p.PaymentStatus != "unpaid" || len(w.Commerce.Deposits) != 0 || len(w.Commerce.Entries) != 0 || len(w.Commerce.ShippingReservations) != 2 {
		t.Fatal("purchase/groups did not precede funding", p)
	}
	for _, o := range w.Orders {
		if o.CheckoutID != p.ID || o.ShippingTerms == nil || o.Payment != "unpaid" || o.OrderStatus != "placed" {
			t.Fatal("missing placed shipment snapshot", o)
		}
	}
	reopened, err := reviewpostgres.New(ctx, db, backoffice.EmptyWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	buyer = loginAt(t, h, "storefront", "09000000005")["token"].(string)
	replay := commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", in, "place-before-funding", 200)
	if !bytes.Equal(first, replay) {
		t.Fatal("lost response replay after commit/restart")
	}
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", bad, "place-before-funding", 409)
	w, _ = reopened.Snapshot(ctx)
	if len(w.Commerce.Purchases) != 1 || len(w.Orders) != 2 || len(w.Commerce.Entries) != 0 || len(w.Commerce.Deposits) != 0 {
		t.Fatal("replay duplicated order/funding")
	}
	for _, offer := range w.Offers {
		if offer.Reserved != 1 || offer.Stock != 10 {
			t.Fatal("replay reserved or consumed twice", offer)
		}
	}
}
