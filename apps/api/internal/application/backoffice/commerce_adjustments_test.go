package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func couponDraft() backoffice.CouponInput {
	return backoffice.CouponInput{Code: "TEST10", Name: "تخفیف آزمون مستقل", Active: true, Kind: "percent", Value: "10", Target: "items", AllCategories: true, MinToman: "0", MaxToman: "0", CapToman: "0", GlobalUnlimited: true, UserUnlimited: true, PlatformBPS: 5000}
}
func (h *commerceHarness) coupon(t *testing.T, in backoffice.CouponInput) backoffice.CouponView {
	t.Helper()
	b, e := h.s.SaveCoupon(context.Background(), h.admin, in, h.key())
	return decodeCommerce[backoffice.CouponView](t, b, e)
}
func (h *commerceHarness) v2Input(t *testing.T, a identity.User) backoffice.QuoteInput {
	t.Helper()
	cart := h.cart(t, a)
	addr := h.address(t, a)
	shipping := map[string]string{}
	for _, l := range cart.Items {
		shipping[l.StoreID] = "post"
	}
	return backoffice.QuoteInput{ContractVersion: 2, CartVersion: cart.Version, AddressID: addr.ID, Shipping: shipping, PaymentMethodID: "card-transfer", WalletToman: "0"}
}
func (h *commerceHarness) v2Quote(t *testing.T, a identity.User, in backoffice.QuoteInput) backoffice.QuoteView {
	t.Helper()
	b, e := h.s.CreateQuote(context.Background(), a, in, h.key())
	return decodeCommerce[backoffice.QuoteView](t, b, e)
}
func (h *commerceHarness) v2Accept(t *testing.T, a identity.User, q backoffice.QuoteView) backoffice.PurchaseView {
	t.Helper()
	b, e := h.s.AcceptQuote(context.Background(), a, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, h.key())
	return decodeCommerce[backoffice.PurchaseView](t, b, e)
}
func TestCouponGoalsCategoriesRangesAndFunding(t *testing.T) {
	for _, tc := range []struct{ target, kind, value, discount, total string }{{"items", "percent", "10", "10000", "110000"}, {"shipping", "percent", "10", "2000", "118000"}, {"order", "percent", "10", "12000", "108000"}, {"items", "fixed", "150000", "100000", "20000"}, {"shipping", "fixed", "150000", "20000", "100000"}, {"order", "fixed", "150000", "120000", "0"}} {
		t.Run(tc.target+tc.kind, func(t *testing.T) {
			h := newCommerceHarness(t)
			rule := couponDraft()
			rule.Target = tc.target
			rule.Kind = tc.kind
			rule.Value = tc.value
			h.coupon(t, rule)
			h.add(t, h.buyer, "offer1", 1)
			in := h.v2Input(t, h.buyer)
			in.CouponCode = "test10"
			if tc.total == "0" {
				in.PaymentMethodID = ""
			}
			q := h.v2Quote(t, h.buyer, in)
			if q.DiscountToman != tc.discount || q.TotalToman != tc.total {
				t.Fatal(q)
			}
			w, _ := h.repo.Snapshot(context.Background())
			snap := w.Commerce.Quotes[len(w.Commerce.Quotes)-1]
			var sum, platform, seller commerce.Amount
			for _, a := range snap.Discount.Allocations {
				sum += a.Amount
				platform += a.Platform
				seller += a.Seller
			}
			if sum != snap.Discount.Amount || platform+seller != sum || platform != seller {
				t.Fatal("funding sums")
			}
			dir, e := h.s.ReadCoupons(context.Background(), h.admin)
			if e != nil || dir.Items[0].Reserved != 0 || dir.Items[0].Consumed != 0 {
				t.Fatal("preview consumed", dir, e)
			}
		})
	}
	t.Run("ancestor-range-cap-and-seller-shipping", func(t *testing.T) {
		h := newCommerceHarness(t)
		rule := couponDraft()
		rule.AllCategories = false
		rule.Categories = []string{"gold"}
		rule.Target = "order"
		rule.MinToman = "120000"
		rule.MaxToman = "120000"
		rule.CapToman = "1000"
		h.coupon(t, rule)
		h.add(t, h.buyer, "gold-offer", 1)
		h.add(t, h.buyer, "offer2", 1)
		in := h.v2Input(t, h.buyer)
		in.CouponCode = rule.Code
		q := h.v2Quote(t, h.buyer, in)
		if q.DiscountToman != "1000" || q.TotalToman != "259000" {
			t.Fatal(q)
		}
		in.CouponCode = "missing"
		if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrCoupon) {
			t.Fatal(e)
		}
	})
}
func TestCouponValidationAccessVersionsAndSnapshot(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := couponDraft()
	if _, e := h.s.SaveCoupon(ctx, h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal(e)
	}
	if _, e := h.s.SaveCoupon(ctx, h.finance, in, h.key()); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal(e)
	}
	for _, edit := range []func(*backoffice.CouponInput){func(c *backoffice.CouponInput) { c.Value = "100.01" }, func(c *backoffice.CouponInput) { c.MinToman = "100"; c.MaxToman = "10" }, func(c *backoffice.CouponInput) { c.PlatformBPS = 10001 }, func(c *backoffice.CouponInput) { c.GlobalUnlimited = false }, func(c *backoffice.CouponInput) { c.AllCategories = false }, func(c *backoffice.CouponInput) {
		end := h.clock.Now()
		start := end.Add(time.Hour)
		c.StartsAt = &start
		c.EndsAt = &end
	}} {
		bad := in
		edit(&bad)
		if _, e := h.s.SaveCoupon(ctx, h.admin, bad, h.key()); !errors.Is(e, backoffice.ErrInvalid) {
			t.Fatalf("invalid accepted: %v", e)
		}
	}
	saved := h.coupon(t, in)
	h.add(t, h.buyer, "offer1", 1)
	quoteIn := h.v2Input(t, h.buyer)
	quoteIn.CouponCode = saved.Code
	q := h.v2Quote(t, h.buyer, quoteIn)
	in = saved.CouponInput
	in.Value = "20"
	h.coupon(t, in)
	if _, e := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, h.key()); !errors.Is(e, backoffice.ErrCheckoutChanged) {
		t.Fatal("stale terms", e)
	}
	if _, e := h.s.SaveCoupon(ctx, h.admin, in, h.key()); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("stale version", e)
	}
	q = h.v2Quote(t, h.buyer, quoteIn)
	p := h.v2Accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "card-transfer", q.TotalToman, "12345001"))
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	before, _ := json.Marshal(p.Quote)
	in.Version++
	in.Value = "90"
	h.coupon(t, in)
	all, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	after, _ := json.Marshal(all[0].Quote)
	if string(before) != string(after) {
		t.Fatal("confirmed snapshot changed")
	}
}
func TestCouponQuotaConcurrentReservationExpiryAndCancellation(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	rule := couponDraft()
	rule.GlobalUnlimited = false
	rule.GlobalLimit = 1
	h.coupon(t, rule)
	quotes := []backoffice.QuoteView{}
	actors := []identity.User{h.buyer, h.other}
	for _, a := range actors {
		h.add(t, a, "offer1", 1)
		in := h.v2Input(t, a)
		in.CouponCode = rule.Code
		quotes = append(quotes, h.v2Quote(t, a, in))
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := range actors {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q := quotes[i]
			_, e := h.s.AcceptQuote(ctx, actors[i], backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, fmt.Sprintf("quota-%d", i))
			results <- e
		}(i)
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else if !errors.Is(e, backoffice.ErrCoupon) {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatal(success)
	}
	dir, _ := h.s.ReadCoupons(ctx, h.admin)
	if dir.Items[0].Reserved != 1 {
		t.Fatal(dir)
	}
	h.clock.Advance(31 * time.Minute)
	dir, _ = h.s.ReadCoupons(ctx, h.admin)
	if dir.Items[0].Reserved != 0 {
		t.Fatal("expired quota held")
	}
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.CouponCode = rule.Code
	p := h.v2Accept(t, h.buyer, h.v2Quote(t, h.buyer, in))
	if _, e := h.s.CancelPurchase(ctx, h.buyer, p.ID, "لغو آزمایشی", h.key()); e != nil {
		t.Fatal(e)
	}
	dir, _ = h.s.ReadCoupons(ctx, h.admin)
	if dir.Items[0].Reserved != 0 {
		t.Fatal("cancel quota held")
	}
}
func TestSelectedWalletEarmarkNoDoubleDebitAndLegacyIsolation(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "22345001"))
	// Recreate the pre-policy fixture: a cash ledger balance without source metadata.
	// Only this isolated memory fixture is changed; no ledger entries are removed.
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.CreditAccounts = nil
		w.Commerce.CreditLots = nil
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "20000"
	q := h.v2Quote(t, h.buyer, in)
	p := h.v2Accept(t, h.buyer, q)
	wallet := h.wallet(t, h.buyer)
	if wallet.HeldToman != "20000" || wallet.AvailableToman != "180000" || p.ShortfallToman != "100000" {
		t.Fatal(wallet, p)
	}
	// The unselected180000 must not fund the order automatically.
	if _, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key()); !errors.Is(e, backoffice.ErrInsufficientWallet) {
		t.Fatal(e)
	}
	d := h.deposit(t, h.buyer, p.ID, "card-transfer", "110000", "22345002")
	h.approve(t, d)
	wallet = h.wallet(t, h.buyer)
	if wallet.HeldToman != "120000" || wallet.AvailableToman != "190000" {
		t.Fatal(wallet)
	}
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	if _, e = h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key()); e != nil {
		t.Fatal(e)
	}
	w, _ := h.repo.Snapshot(ctx)
	debits := 0
	for _, entry := range w.Commerce.Entries {
		if entry.Kind == "order-payment" {
			debits++
		}
	}
	if debits != 1 || len(w.Commerce.CreditAccounts) != 1 || w.Commerce.CreditLots[0].Source != "legacy-unattributed" {
		t.Fatal("allocation metadata/ledger", w.Commerce)
	}
	wallet = h.wallet(t, h.buyer)
	if wallet.BalanceToman != "190000" || wallet.HeldToman != "0" {
		t.Fatal(wallet)
	}
	if _, e := h.s.AdminCancelPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "لغو آزمایشی"}, h.key()); e != nil {
		t.Fatal(e)
	}
	if h.wallet(t, h.buyer).BalanceToman != "310000" {
		t.Fatal("snapshot refund")
	}
}
func TestWalletCoverageLimitsExpiryAndFreeOrder(t *testing.T) {
	t.Run("full-selected", func(t *testing.T) {
		h := newCommerceHarness(t)
		h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "32345001"))
		h.add(t, h.buyer, "offer1", 1)
		in := h.v2Input(t, h.buyer)
		in.PaymentMethodID = ""
		in.WalletToman = "120000"
		p := h.v2Accept(t, h.buyer, h.v2Quote(t, h.buyer, in))
		b, e := h.s.ConfirmPurchase(context.Background(), h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
		decodeCommerce[backoffice.PurchaseView](t, b, e)
		if h.wallet(t, h.buyer).BalanceToman != "80000" {
			t.Fatal("wrong selected debit")
		}
	})
	t.Run("expiry-release-late-credit", func(t *testing.T) {
		h := newCommerceHarness(t)
		h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "32345002"))
		h.add(t, h.buyer, "offer1", 1)
		in := h.v2Input(t, h.buyer)
		in.WalletToman = "20000"
		p := h.v2Accept(t, h.buyer, h.v2Quote(t, h.buyer, in))
		d := h.deposit(t, h.buyer, p.ID, "card-transfer", "100000", "32345003")
		h.clock.Advance(31 * time.Minute)
		h.approve(t, d)
		w := h.wallet(t, h.buyer)
		if w.HeldToman != "0" || w.AvailableToman != "300000" {
			t.Fatal(w)
		}
	})
	t.Run("free-manual-confirm-consumes-quota", func(t *testing.T) {
		h := newCommerceHarness(t)
		rule := couponDraft()
		rule.Kind = "fixed"
		rule.Value = "999999"
		rule.Target = "order"
		rule.GlobalUnlimited = false
		rule.GlobalLimit = 1
		h.coupon(t, rule)
		h.add(t, h.buyer, "offer1", 1)
		in := h.v2Input(t, h.buyer)
		in.CouponCode = rule.Code
		in.PaymentMethodID = ""
		q := h.v2Quote(t, h.buyer, in)
		p := h.v2Accept(t, h.buyer, q)
		if p.Status == "confirmed" {
			t.Fatal("automatic confirmation")
		}
		b, e := h.s.ConfirmPurchase(context.Background(), h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
		p = decodeCommerce[backoffice.PurchaseView](t, b, e)
		if _, e := h.s.AdminCancelPurchase(context.Background(), h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "بازگشت آزمایشی"}, h.key()); e != nil {
			t.Fatal(e)
		}
		w, _ := h.repo.Snapshot(context.Background())
		if len(w.Commerce.Entries) != 0 {
			t.Fatal("zero ledger movement")
		}
		dir, _ := h.s.ReadCoupons(context.Background(), h.admin)
		if dir.Items[0].Consumed != 1 {
			t.Fatal("quota restored after confirmation")
		}
	})
	t.Run("wallet-and-funding-rules-on-full-order", func(t *testing.T) {
		h := newCommerceHarness(t)
		h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "32345004"))
		h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
			for i := range w.Commerce.Settings.Methods {
				if w.Commerce.Settings.Methods[i].Kind == "wallet" {
					w.Commerce.Settings.Methods[i].ExcludedCategories = []string{"gold"}
				}
			}
			return nil
		})
		h.add(t, h.buyer, "gold-offer", 1)
		in := h.v2Input(t, h.buyer)
		in.WalletToman = "120000"
		in.PaymentMethodID = ""
		if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrInvalidMethod) {
			t.Fatal(e)
		}
		in.WalletToman = "0"
		in.PaymentMethodID = "local"
		if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrInvalidMethod) {
			t.Fatal(e)
		}
	})
}
func TestWarrantyMappingDoesNotAlterRealVariantOrHistoricalItem(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Attributes = []catalog.AttributeDefinition{{ID: "warranty", Code: "warranty", Name: "گارانتی", Kind: "enum", Options: []string{"شرکت الف ۶ماه", "شرکت ب ۶ماه"}, Version: 1, Groups: []catalog.AttributeValueGroup{{ID: "company-six", Name: "شش ماه شرکتی", Values: []string{"شرکت الف ۶ماه", "شرکت ب ۶ماه"}, Active: true}}}}
		w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "warranty", Axis: true}}
		w.Variants[0].Values = []catalog.AttributeValue{{AttributeID: "warranty", Value: "شرکت الف ۶ماه"}}
		w.Variants = append(w.Variants, catalog.Variant{ID: "phone-b", ProductID: "phone", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "warranty", Value: "شرکت ب ۶ماه"}}})
		w.Offers = append(w.Offers, catalog.Offer{ID: "better", ProductID: "phone", VariantID: "phone-b", StoreID: "store2", Price: 900000, Stock: 3, Version: 1})
		return nil
	})
	w, _ := h.repo.Snapshot(ctx)
	detail, e := backoffice.BuildProductDetail(w, "phone", "phone-v", "")
	if e != nil || detail.SelectedVariantID != "phone-v" {
		t.Fatal("old raw URL", detail, e)
	}
	detail, e = backoffice.BuildProductDetail(w, "phone", "phone-v", "", true)
	if e != nil || detail.SelectedVariantID != "phone-b" || detail.SelectedOfferID != "better" || len(detail.Axes[0].Groups) != 1 {
		t.Fatal(detail, e)
	}
	detail, e = backoffice.BuildProductDetail(w, "phone", "phone-b", "offer1")
	if e != nil || detail.SelectedVariantID != "phone-v" {
		t.Fatal("explicit seller", detail, e)
	}
	if _, e = h.s.ValidateSelection(ctx, "phone", backoffice.ProductSelectionInput{VariantID: "phone-b", OfferID: "offer1", Quantity: 1}); e == nil {
		t.Fatal("mismatched real variant accepted")
	}
	h.add(t, h.buyer, "better", 1)
	in := h.v2Input(t, h.buyer)
	q := h.v2Quote(t, h.buyer, in)
	p := h.v2Accept(t, h.buyer, q)
	w, _ = h.repo.Snapshot(ctx)
	original := w.Orders[0].Items[0].VariantLabel
	a := w.Attributes[0]
	a.Groups[0].Name = "عنوان ویرایش‌شده"
	if e := h.s.Execute(ctx, h.admin, backoffice.Command{Action: "attribute.update", Target: a.ID, Version: a.Version, Attribute: &a, IdempotencyKey: h.key()}); e != nil {
		t.Fatal(e)
	}
	w, _ = h.repo.Snapshot(ctx)
	if w.Orders[0].Items[0].VariantID != "phone-b" || original != "شرکت ب ۶ماه" || w.Orders[0].Items[0].VariantLabel != original || p.Quote.Lines[0].VariantLabel != original {
		t.Fatal("real warranty snapshot lost")
	}
}

func TestWalletConcurrentHoldAndRefundCannotStealCredit(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "42345001"))
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "120000"
	in.PaymentMethodID = ""
	a := h.v2Quote(t, h.buyer, in)
	b := h.v2Quote(t, h.buyer, in)
	var wg sync.WaitGroup
	ch := make(chan error, 2)
	for i, q := range []backoffice.QuoteView{a, b} {
		wg.Add(1)
		go func(i int, q backoffice.QuoteView) {
			defer wg.Done()
			_, e := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, fmt.Sprintf("credit-hold-%d", i))
			ch <- e
		}(i, q)
	}
	wg.Wait()
	close(ch)
	success := 0
	for e := range ch {
		if e == nil {
			success++
		} else if !errors.Is(e, backoffice.ErrCheckoutChanged) && !errors.Is(e, backoffice.ErrInsufficientWallet) {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatal(success)
	}
	if h.wallet(t, h.buyer).HeldToman != "120000" {
		t.Fatal("hold duplicated")
	}
	// Existing legacy full-wallet writer also must not spend this reserved amount.
	h.add(t, h.buyer, "offer1", 1)
	legacy := h.quote(t, h.buyer, "wallet")
	p := h.accept(t, h.buyer, legacy)
	if _, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key()); !errors.Is(e, backoffice.ErrInsufficientWallet) {
		t.Fatal("legacy stole credit", e)
	}
	h.clock.Advance(31 * time.Minute)
	if h.wallet(t, h.buyer).HeldToman != "0" {
		t.Fatal("wallet read did not expire holds")
	}
}
func TestWarrantyGroupsValidationAndPermission(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	a := catalog.AttributeDefinition{Code: "test-warranty-groups", Name: "گارانتی آزمون", Kind: "enum", Options: []string{"الف", "ب"}, Groups: []catalog.AttributeValueGroup{{ID: "six-month", Name: "شش ماهه", Values: []string{"الف", "ب"}, Active: true}}}
	cmd := backoffice.Command{Action: "attribute.create", Attribute: &a, IdempotencyKey: h.key()}
	if e := h.s.Execute(ctx, h.finance, cmd); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal(e)
	}
	if e := h.s.Execute(ctx, h.admin, cmd); e != nil {
		t.Fatal(e)
	}
	for _, bad := range []catalog.AttributeValueGroup{{ID: "unknown", Name: "ناشناخته", Values: []string{"خارج از تعریف"}, Active: true}, {ID: "overlap", Name: "تکراری", Values: []string{"الف"}, Active: true}} {
		w, _ := h.repo.Snapshot(ctx)
		var saved catalog.AttributeDefinition
		for _, v := range w.Attributes {
			if v.Code == a.Code {
				saved = v
			}
		}
		saved.Groups = append(saved.Groups, bad)
		if e := h.s.Execute(ctx, h.admin, backoffice.Command{Action: "attribute.update", Target: saved.ID, Version: saved.Version, Attribute: &saved, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrInvalid) {
			t.Fatal("bad mapping", e)
		}
	}
}

func TestGroupOnlyMappingSupportsImportedLargeDefinitionAndChecksVersion(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	options := []string{}
	for i := 0; i < 150; i++ {
		options = append(options, fmt.Sprintf("شرکت آزمون %d", i))
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Attributes = append(w.Attributes, catalog.AttributeDefinition{ID: "large-warranty", Code: "large-warranty", Name: "گارانتی واردشده", Kind: "enum", Options: options, Version: 1})
		return nil
	})
	input := catalog.AttributeDefinition{Groups: []catalog.AttributeValueGroup{{ID: "six-month", Name: "شش ماهه شرکتی", Active: true, Values: options}}}
	command := backoffice.Command{Action: "attribute.groups.set", Target: "large-warranty", Attribute: &input, Version: 1, IdempotencyKey: h.key()}
	if e := h.s.Execute(ctx, h.finance, command); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal(e)
	}
	if e := h.s.Execute(ctx, h.admin, command); e != nil {
		t.Fatal(e)
	}
	if e := h.s.Execute(ctx, h.admin, command); e != nil {
		t.Fatal("retry", e)
	}
	command.IdempotencyKey = h.key()
	if e := h.s.Execute(ctx, h.admin, command); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("version", e)
	}
	w, _ := h.repo.Snapshot(ctx)
	for _, a := range w.Attributes {
		if a.ID == "large-warranty" && (len(a.Options) != 150 || len(a.Groups) != 1 || a.Version != 2) {
			t.Fatal(a)
		}
	}
}

func TestCouponTimeRangeAndUserQuotaBoundaries(t *testing.T) {
	for _, tc := range []string{"inactive", "future", "expired", "minimum", "maximum", "wrong-category"} {
		t.Run(tc, func(t *testing.T) {
			h := newCommerceHarness(t)
			rule := couponDraft()
			switch tc {
			case "inactive":
				rule.Active = false
			case "future":
				at := h.clock.Now().Add(time.Second)
				rule.StartsAt = &at
			case "expired":
				at := h.clock.Now()
				rule.EndsAt = &at
			case "minimum":
				rule.MinToman = "100000.1"
			case "maximum":
				rule.MaxToman = "99999.9"
			case "wrong-category":
				rule.AllCategories = false
				rule.Categories = []string{"gold"}
			}
			h.coupon(t, rule)
			h.add(t, h.buyer, "offer1", 1)
			in := h.v2Input(t, h.buyer)
			in.CouponCode = rule.Code
			if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrCoupon) {
				t.Fatal(e)
			}
		})
	}
	h := newCommerceHarness(t)
	rule := couponDraft()
	rule.Target = "order"
	rule.Value = "100"
	rule.UserUnlimited = false
	rule.UserLimit = 1
	h.coupon(t, rule)
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.CouponCode = rule.Code
	in.PaymentMethodID = ""
	p := h.v2Accept(t, h.buyer, h.v2Quote(t, h.buyer, in))
	if _, e := h.s.ConfirmPurchase(context.Background(), h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key()); e != nil {
		t.Fatal(e)
	}
	h.add(t, h.buyer, "offer1", 1)
	in = h.v2Input(t, h.buyer)
	in.CouponCode = rule.Code
	in.PaymentMethodID = ""
	if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrCoupon) {
		t.Fatal("per-user quota", e)
	}
	h.add(t, h.other, "offer1", 1)
	in = h.v2Input(t, h.other)
	in.CouponCode = rule.Code
	in.PaymentMethodID = ""
	h.v2Quote(t, h.other, in)
}
func TestSelectedWalletDoesNotBypassFundingOrderCap(t *testing.T) {
	h := newCommerceHarness(t)
	h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "6000000", "52345001"))
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Price = 200000000; w.Offers[0].Version++; return nil })
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "6000000"
	if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); !errors.Is(e, backoffice.ErrInvalidMethod) {
		t.Fatal("remaining funding below15m bypassed full20m cap", e)
	}
}
