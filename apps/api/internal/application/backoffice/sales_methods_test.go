package backoffice_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"sync"
	"testing"
)

func salesFixture(t *testing.T) (*commerceHarness, commerce.SalesSettings) {
	h := newCommerceHarness(t)
	d, e := h.s.ReadSales(context.Background(), h.admin)
	if e != nil {
		t.Fatal(e)
	}
	in := d.SalesSettings
	in.Methods = []commerce.SalesMethod{{ID: "cash", Name: "نقدی", Active: true, PaymentMethodIDs: []string{"card-transfer"}}, {ID: "credit", Name: "اعتباری آزمایشی", Active: true, MarkupBPS: 400, PaymentMethodIDs: []string{"local"}}}
	return h, in
}
func saveSales(t *testing.T, h *commerceHarness, in commerce.SalesSettings) commerce.SalesSettings {
	t.Helper()
	raw, e := h.s.SaveSales(context.Background(), h.admin, in, h.key())
	return decodeCommerce[commerce.SalesSettings](t, raw, e)
}
func TestSalesMethodsRateValidationPermissionsConcurrencyAndInheritance(t *testing.T) {
	ctx := context.Background()
	h, in := salesFixture(t)
	if _, e := h.s.SaveSales(ctx, h.ops, in, h.key()); e != backoffice.ErrForbidden {
		t.Fatal("financial permission", e)
	}
	for _, rate := range []int{-1, 10001} {
		bad := in
		bad.Methods = append([]commerce.SalesMethod{}, in.Methods...)
		bad.Methods[1].MarkupBPS = rate
		if _, e := h.s.SaveSales(ctx, h.admin, bad, h.key()); e != backoffice.ErrInvalid {
			t.Fatal("rate", e)
		}
	}
	bad := in
	bad.Methods = append([]commerce.SalesMethod{}, in.Methods...)
	bad.Methods[1].PaymentMethodIDs = []string{"card-transfer"}
	if _, e := h.s.SaveSales(ctx, h.admin, bad, h.key()); e != backoffice.ErrInvalid {
		t.Fatal("duplicate mapping", e)
	}
	in = saveSales(t, h, in)
	var wg sync.WaitGroup
	result := make(chan error, 2)
	for _, key := range []string{"sales-a", "sales-b"} {
		wg.Add(1)
		go func(k string) { defer wg.Done(); _, e := h.s.SaveSales(ctx, h.admin, in, k); result <- e }(key)
	}
	wg.Wait()
	close(result)
	ok, conflict := 0, 0
	for e := range result {
		if e == nil {
			ok++
		} else if e == backoffice.ErrConflict {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatal("concurrent write")
	}
	d, _ := h.s.ReadSales(ctx, h.admin)
	in = d.SalesSettings
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Categories = append(w.Categories, catalog.Category{ID: "child", Name: "فرزند", ParentID: w.Products[0].CategoryID, Published: true})
		return nil
	})
	w, _ := h.repo.Snapshot(ctx)
	parent := w.Products[0].CategoryID
	in.Overrides = []commerce.CategoryMarkup{{CategoryID: parent, MethodID: "credit", MarkupBPS: 600}, {CategoryID: "child", MethodID: "credit", MarkupBPS: 0}}
	in = saveSales(t, h, in)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Products[0].CategoryID = "child"; return nil })
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	if q.MarkupToman != "0" || q.Lines[0].Markup.CategoryID != "child" {
		t.Fatal("explicit zero")
	}
	in.Overrides = in.Overrides[:1]
	saveSales(t, h, in)
	q = h.quote(t, h.buyer, "local")
	if q.Lines[0].Markup.BPS != 600 || q.Lines[0].Markup.CategoryID != parent {
		t.Fatal("nearest ancestor")
	}
}
func TestSalesMarkupCouponWalletAndImmutableAcceptedPrice(t *testing.T) {
	ctx := context.Background()
	h, in := salesFixture(t)
	in = saveSales(t, h, in)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Offers {
			if w.Offers[i].ID == "offer1" {
				w.Offers[i].Price = 10000
				w.Offers[i].RegularPrice = 12000
			}
		}
		for i := range w.Commerce.Settings.Shipping {
			w.Commerce.Settings.Shipping[i].Fee = 0
		}
		return nil
	})
	cart := h.add(t, h.buyer, "offer1", 1)
	address := h.address(t, h.buyer)
	raw, e := h.s.SaveCoupon(ctx, h.admin, backoffice.CouponInput{Code: "SALE10", Name: "تست", Active: true, Kind: "percent", Value: "10", Target: "items", AllCategories: true, MinToman: "0", MaxToman: "0", CapToman: "0", GlobalUnlimited: true, UserUnlimited: true, PlatformBPS: 10000}, h.key())
	if e != nil {
		t.Fatal(e)
	}
	_ = raw
	input := backoffice.QuoteInput{ContractVersion: 2, CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "local", CouponCode: "SALE10", WalletToman: "0"}
	raw, e = h.s.CreateQuote(ctx, h.buyer, input, h.key())
	q := decodeCommerce[backoffice.QuoteView](t, raw, e)
	if q.TotalToman != "936" || q.MarkupToman != "40" || q.DiscountToman != "104" || q.CashSubtotalToman != "1000" || q.DirectDiscountToman != "200" {
		t.Fatalf("pricing %#v", q)
	}
	preview, e := h.s.PreviewCheckout(ctx, h.buyer, input)
	if e != nil || preview.Quote.TotalToman != "936" {
		t.Fatal("preview", e)
	}
	// A policy change between quote and consent must be rejected.
	in.Methods[1].MarkupBPS = 800
	in = saveSales(t, h, in)
	if _, e = h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, h.key()); e != backoffice.ErrCheckoutChanged {
		t.Fatal("stale consent", e)
	}
	in.Methods[1].MarkupBPS = 400
	in = saveSales(t, h, in)
	raw, e = h.s.CreateQuote(ctx, h.buyer, input, h.key())
	q = decodeCommerce[backoffice.QuoteView](t, raw, e)
	raw, e = h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, h.key())
	p := decodeCommerce[backoffice.PurchaseView](t, raw, e)
	before, _ := json.Marshal(p.Quote.Lines)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", "936", "sales-price-paid"))
	// Rate changes after acceptance do not reprice the immutable purchase.
	in.Methods[1].MarkupBPS = 900
	saveSales(t, h, in)
	current, _ := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	raw, e = h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: current[0].Version}, h.key())
	confirmed := decodeCommerce[backoffice.PurchaseView](t, raw, e)
	if confirmed.Quote.TotalToman != "936" || confirmed.Status != "confirmed" {
		t.Fatal("frozen accepted price")
	}
	after, _ := json.Marshal(current[0].Quote.Lines)
	if string(before) != string(after) {
		t.Fatal("historical snapshot rewritten")
	}
	if h.wallet(t, h.buyer).BalanceToman != "0" {
		t.Fatal("funding/payment mismatch")
	}
}

func TestSalesFullWalletCannotBypassRevokedSelectedMethod(t *testing.T) {
	for _, reason := range []string{"sales-disabled", "payment-disabled", "payment-limit"} {
		t.Run(reason, func(t *testing.T) {
			ctx := context.Background()
			h, in := salesFixture(t)
			in = saveSales(t, h, in)
			h.credit(t, "124000", false)
			h.add(t, h.buyer, "offer1", 1)
			input := h.v2Input(t, h.buyer)
			input.PaymentMethodID = "local"
			input.WalletToman = "124000"
			q := h.v2Quote(t, h.buyer, input)
			p := h.v2Accept(t, h.buyer, q)
			if q.FundingToman != "0" {
				t.Fatal("fixture must be fully funded")
			}
			if reason == "sales-disabled" {
				in.Methods[1].Active = false
				saveSales(t, h, in)
			} else {
				if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
					for i := range w.Commerce.Settings.Methods {
						if w.Commerce.Settings.Methods[i].ID == "local" {
							if reason == "payment-disabled" {
								w.Commerce.Settings.Methods[i].Active = false
							} else {
								w.Commerce.Settings.Methods[i].Max = 1
							}
						}
					}
					return nil
				}); e != nil {
					t.Fatal(e)
				}
			}
			raw, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
			result := decodeCommerce[backoffice.PurchaseView](t, raw, e)
			if result.Status != "requires-reconfirmation" {
				t.Fatal("selected method bypassed by full wallet", result.Status)
			}
			if h.wallet(t, h.buyer).AvailableToman != "124000" {
				t.Fatal("reserved credit was not released")
			}
		})
	}
}

func TestSalesMixedCategoriesUseIndependentRatesAndRespectLimits(t *testing.T) {
	h, in := salesFixture(t)
	in.Overrides = []commerce.CategoryMarkup{{CategoryID: "regular", MethodID: "credit", MarkupBPS: 200}, {CategoryID: "gold", MethodID: "credit", MarkupBPS: 600}}
	saveSales(t, h, in)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "gold-offer", 1)
	// Local funding excludes gold, including its descendants, even when the sale method is active.
	if _, e := h.s.CreateQuote(context.Background(), h.buyer, backoffice.QuoteInput{CartVersion: h.cart(t, h.buyer).Version, AddressID: h.address(t, h.buyer).ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "local"}, h.key()); e != backoffice.ErrInvalidMethod {
		t.Fatal("category restriction", e)
	}
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := range w.Commerce.Settings.Methods {
			w.Commerce.Settings.Methods[i].ExcludedCategories = nil
		}
		return nil
	})
	q := h.quote(t, h.buyer, "local")
	if q.MarkupToman != "8000" || q.CashSubtotalToman != "200000" || len(q.Lines) != 2 {
		t.Fatal("mixed category calculation", q)
	}
	rates := map[string]int{}
	for _, l := range q.Lines {
		rates[l.ProductID] = l.Markup.BPS
	}
	if rates["phone"] != 200 || rates["gold-product"] != 600 {
		t.Fatal("separate inherited rates", rates)
	}
}
