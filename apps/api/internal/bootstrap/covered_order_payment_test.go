package bootstrap_test

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

type rejectCoveredPaymentCommit struct{ backoffice.Repository }

func (r rejectCoveredPaymentCommit) Update(ctx context.Context, apply func(*backoffice.Workspace) error) error {
	return r.Repository.Update(ctx, func(w *backoffice.Workspace) error {
		if e := apply(w); e != nil {
			return e
		}
		return errors.New("isolated late persistence failure")
	})
}

func TestCoveredOrderPostgresAtomicPaymentRollbackAndRestartReplay(t *testing.T) {
	for _, mode := range []string{"wallet", "free"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			repo, db := cachePostgres(t)
			seed := commerceSeed()
			seed.Offers[0].RegularPrice = 1200000
			seed.Offers = append(seed.Offers, catalog.Offer{ID: "second-offer", ProductID: "product", VariantID: "variant", StoreID: "s2", Price: 1200000, RegularPrice: 1500000, Stock: 10, Version: 1})
			if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
				t.Fatal(e)
			}
			h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
			admin := login(t, h, "09000000001")
			buyer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
			commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/wallet/credits", backoffice.AdminCreditInput{UserID: "viewer", AmountToman: "500000", Title: "اعتبار ساختگی", Description: "آزمون اتمیک پرداخت", Withdrawable: false}, "credit", 200)
			cart := readJSON[backoffice.CartView](t, commerceHTTP(t, h, buyer, "", "GET", "/storefront/cart", nil, "", 200))
			for _, offer := range []string{"offer", "second-offer"} {
				cart = readJSON[backoffice.CartView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: offer, Quantity: 1, Version: cart.Version}, "add-"+offer, 200))
			}
			address := readJSON[commerce.Address](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", backoffice.AddressInput{Recipient: "گیرنده ساختگی", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Phone: "09000000111", Province: "تهران", City: "تهران", Street: "نشانی ساختگی آزمون", PostalCode: "1234567890", HouseNumber: "12"}, "address", 200))
			in := backoffice.QuoteInput{ContractVersion: 2, CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post", "s2": "post"}, WalletToman: "999999"}
			if mode == "free" {
				coupon := backoffice.CouponInput{Code: "FREE-ORDER", Name: "سفارش رایگان ساختگی", Active: true, Kind: "percent", Value: "100", Target: "order", AllCategories: true, MinToman: "0", MaxToman: "0", CapToman: "0", GlobalUnlimited: true, UserUnlimited: true, PlatformBPS: 10000}
				commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/coupons", coupon, "coupon", 200)
				in.CouponCode = coupon.Code
			}
			q := readJSON[backoffice.QuoteView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/quotes", in, "quote", 200))
			if q.ReferenceSubtotalToman != "270000" || q.DirectDiscountToman != "50000" {
				t.Fatal("reference subtotal must precede item discounts", q)
			}
			placement := backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}
			before, _ := repo.Snapshot(ctx)
			failed := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, rejectCoveredPaymentCommit{repo})
			failedBuyer := loginAt(t, failed, "storefront", "09000000005")["token"].(string)
			commerceHTTP(t, failed, failedBuyer, "", "POST", "/storefront/checkout/purchases", placement, "place-covered", 503)
			after, _ := repo.Snapshot(ctx)
			if !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Offers, after.Offers) {
				t.Fatal("failed commit left order, debit, stock or capacity")
			}
			first := commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", placement, "place-covered", 200)
			p := readJSON[backoffice.PurchaseView](t, first)
			if p.PaymentStatus != "paid" || p.OrderStatus != "placed" || p.Status != "confirmed" || len(p.OrderIDs) != 2 {
				t.Fatal("not automatically paid or processing started", p)
			}
			reopened, e := reviewpostgres.New(ctx, db, backoffice.EmptyWorkspace())
			if e != nil {
				t.Fatal(e)
			}
			h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
			buyer = loginAt(t, h, "storefront", "09000000005")["token"].(string)
			detail := readJSON[backoffice.PurchaseView](t, commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/orders/"+p.ID, nil, "", 200))
			if detail.Quote.ReferenceSubtotalToman != q.ReferenceSubtotalToman {
				t.Fatal("snapshot gross changed across restart")
			}
			replay := commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", placement, "place-covered", 200)
			if !bytes.Equal(first, replay) {
				t.Fatal("lost response replay changed")
			}
			// A new request key for the same quote must also remain a single purchase/payment.
			commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", placement, "place-covered-again", 200)
			state, _ := reopened.Snapshot(ctx)
			debits := 0
			for _, entry := range state.Commerce.Entries {
				if entry.Kind == "order-payment" {
					debits++
					if commerce.Toman(-entry.Amount) != q.TotalToman {
						t.Fatal("incorrect debit")
					}
				}
			}
			if mode == "free" && debits != 0 || mode == "wallet" && debits != 1 || len(state.Commerce.Purchases) != 1 || len(state.Orders) != 2 || len(state.Commerce.Deposits) != 0 {
				t.Fatal("duplicated financial effect", debits)
			}
			for _, offer := range state.Offers {
				if offer.Reserved != 0 || offer.Stock != 9 {
					t.Fatal("stock duplicated", offer)
				}
			}
			for _, o := range state.Orders {
				if o.OrderStatus != "placed" || o.Payment != "paid" || o.Status != "awaiting-supply" {
					t.Fatal("processing started automatically", o)
				}
			}
			paidEvents := 0
			for _, event := range state.Commerce.Purchases[0].Events {
				if event.Action == "order.paid" {
					paidEvents++
				}
			}
			if paidEvents != 1 {
				t.Fatal("paid event duplicated", paidEvents)
			}
		})
	}
}
