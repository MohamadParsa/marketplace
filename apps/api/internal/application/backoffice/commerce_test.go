package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type commerceClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *commerceClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *commerceClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

type commerceIDs struct{ n atomic.Int64 }

func (i *commerceIDs) Token() (string, error) { return fmt.Sprintf("%064d", i.n.Add(1)), nil }

type commerceHarness struct {
	s                                 *backoffice.Service
	repo                              *backofficememory.Repository
	clock                             *commerceClock
	buyer, other, admin, finance, ops identity.User
	seq                               int
}

func newCommerceHarness(t *testing.T) *commerceHarness {
	t.Helper()
	w := backofficememory.Fixture()
	w.Orders = nil
	w.Purchases = nil
	w.Cases = nil
	w.Transfers = nil
	w.Ledger = nil
	w.Categories = []catalog.Category{{ID: "regular", Name: "دیجیتال", Published: true, Version: 1}, {ID: "gold", Name: "طلا", Published: true, Version: 1}, {ID: "coin", Name: "سکه", ParentID: "gold", Published: true, Version: 1}}
	w.Products = []catalog.Product{{ID: "phone", Name: "گوشی نمونه", CategoryID: "regular", Status: "published", Model: "shared", ImageID: "photo", Version: 1}, {ID: "gold-product", Name: "سکه نمونه", CategoryID: "coin", Status: "published", Model: "shared", ImageID: "photo", Version: 1}}
	w.IconMedia = []catalog.IconMedia{{ID: "photo", Data: []byte{1}}}
	w.Stores = []store.Store{{ID: "store1", Name: "فروشگاه یک", Active: true, LeadDays: 1}, {ID: "store2", Name: "فروشگاه دو", Active: true, LeadDays: 3}}
	w.Variants = []catalog.Variant{{ID: "phone-v", ProductID: "phone", Version: 1}, {ID: "gold-v", ProductID: "gold-product", Version: 1}}
	w.Offers = []catalog.Offer{{ID: "offer1", ProductID: "phone", VariantID: "phone-v", StoreID: "store1", Stock: 5, Price: 1000000, Version: 1}, {ID: "offer2", ProductID: "phone", VariantID: "phone-v", StoreID: "store2", Stock: 5, Price: 1200000, Version: 1}, {ID: "gold-offer", ProductID: "gold-product", VariantID: "gold-v", StoreID: "store1", Stock: 5, Price: 1000000, Version: 1}}
	settings := commerce.DefaultSettings()
	settings.Geography = []commerce.Province{{ID: "tehran", Name: "تهران", Active: true, Cities: []commerce.City{{ID: "tehran-city", Name: "تهران", Active: true}}}, {ID: "qom", Name: "قم", Active: true, Cities: []commerce.City{{ID: "qom-city", Name: "قم", Active: true}}}}
	settings.Destination = commerce.Destination{ID: "test-destination", CardNumber: "4242424242424242", Holder: "فروشگاه نمونه", Bank: "بانک آزمایشی"}
	for i := range settings.Methods {
		settings.Methods[i].Active = true
		if settings.Methods[i].ID == "local" {
			settings.Methods[i].ExcludedCategories = []string{"gold"}
			settings.Methods[i].Instructions = "داده آزمون؛ پرداخت حضوری با بررسی تیم"
		}
	}
	capacity := 1000 // Explicit fixture quota; production has no assumed capacity.
	settings.Shipping = []commerce.ShippingMethod{{DailyCapacity: &capacity, ID: "post", Name: "پست نمونه", Active: true, AllCities: true, Fee: 200000, MinDays: 1, MaxDays: 3}}
	w.Commerce = &commerce.State{Settings: settings}
	h := &commerceHarness{clock: &commerceClock{now: time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)}}
	for _, u := range w.Users {
		a := identity.User{ID: u.ID, Active: u.Active, AccessVersion: u.AccessVersion, Surface: identity.SurfaceCustomer}
		if h.buyer.ID == "" {
			h.buyer = a
		} else if h.other.ID == "" {
			h.other = a
		}
		if u.Role == identity.RoleSuperAdmin {
			h.admin, _ = w.BackofficeIdentity(u)
		}
		if u.Role == identity.RoleFinance {
			h.finance, _ = w.BackofficeIdentity(u)
		}
		if u.Role == identity.RoleOperations {
			h.ops, _ = w.BackofficeIdentity(u)
		}
	}
	if h.admin.ID == "" || h.finance.ID == "" || h.buyer.ID == "" {
		t.Fatal("fixture roles")
	}
	h.repo = backofficememory.New(w)
	h.s = backoffice.NewService(h.repo, h.clock, &commerceIDs{})
	return h
}
func (h *commerceHarness) key() string { h.seq++; return fmt.Sprintf("test-%d", h.seq) }
func decodeCommerce[T any](t *testing.T, b json.RawMessage, e error) T {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
	var v T
	if e = json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func (h *commerceHarness) cart(t *testing.T, a identity.User) backoffice.CartView {
	t.Helper()
	v, e := h.s.ReadCart(context.Background(), backoffice.CartOwner{Actor: a})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func (h *commerceHarness) add(t *testing.T, a identity.User, offer string, qty int) backoffice.CartView {
	t.Helper()
	v := h.cart(t, a)
	b, e := h.s.MutateCart(context.Background(), backoffice.CartOwner{Actor: a}, backoffice.CartMutation{Action: "add", OfferID: offer, Quantity: qty, Version: v.Version}, h.key())
	return decodeCommerce[backoffice.CartView](t, b, e)
}
func (h *commerceHarness) address(t *testing.T, a identity.User) commerce.Address {
	t.Helper()
	b, e := h.s.SaveAddress(context.Background(), a, backoffice.AddressInput{HouseNumber: "12", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Recipient: "گیرنده نمونه", Phone: "09000000111", Province: "تهران", City: "تهران", Street: "نشانی ساختگی برای آزمون", PostalCode: "1234567890"}, h.key())
	return decodeCommerce[commerce.Address](t, b, e)
}
func (h *commerceHarness) quote(t *testing.T, a identity.User, method string) backoffice.QuoteView {
	t.Helper()
	cart := h.cart(t, a)
	address := h.address(t, a)
	shipping := map[string]string{}
	for _, l := range cart.Items {
		shipping[l.StoreID] = "post"
	}
	b, e := h.s.CreateQuote(context.Background(), a, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: shipping, PaymentMethodID: method}, h.key())
	return decodeCommerce[backoffice.QuoteView](t, b, e)
}
func (h *commerceHarness) accept(t *testing.T, a identity.User, q backoffice.QuoteView) backoffice.PurchaseView {
	t.Helper()
	b, e := h.s.AcceptQuote(context.Background(), a, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key())
	return decodeCommerce[backoffice.PurchaseView](t, b, e)
}
func (h *commerceHarness) deposit(t *testing.T, a identity.User, purchase, method, amount, ref string) backoffice.DepositView {
	t.Helper()
	b, e := h.s.CreateDeposit(context.Background(), a, backoffice.DepositInput{MethodID: method, PurchaseID: purchase, AmountToman: amount, Date: h.clock.Now().Format("2006-01-02"), Reference: ref}, h.key())
	return decodeCommerce[backoffice.DepositView](t, b, e)
}
func (h *commerceHarness) approve(t *testing.T, d backoffice.DepositView) backoffice.DepositView {
	t.Helper()
	b, e := h.s.ReviewDeposit(context.Background(), h.finance, backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedAmountToman: d.AmountToman, VerifiedDate: d.Date, VerifiedReference: d.Reference}, h.key())
	return decodeCommerce[backoffice.DepositView](t, b, e)
}
func (h *commerceHarness) wallet(t *testing.T, a identity.User) backoffice.WalletView {
	t.Helper()
	v, e := h.s.ReadWallet(context.Background(), a)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestCommerceGuestCartMergeAndOwnerIsolation(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	g, e := h.s.CreateGuestCart(ctx)
	if e != nil {
		t.Fatal(e)
	}
	owner := backoffice.CartOwner{GuestToken: g["cart_token"].(string)}
	v, e := h.s.ReadCart(ctx, owner)
	if e != nil {
		t.Fatal(e)
	}
	b, e := h.s.MutateCart(ctx, owner, backoffice.CartMutation{Action: "add", OfferID: "offer1", Quantity: 2, Version: v.Version}, "guest-add")
	v = decodeCommerce[backoffice.CartView](t, b, e)
	h.add(t, h.buyer, "offer1", 3)
	b, e = h.s.MergeCart(ctx, h.buyer, owner.GuestToken, "merge-once")
	merged := decodeCommerce[backoffice.CartView](t, b, e)
	if merged.Items[0].Quantity != 5 {
		t.Fatal("merge quantity")
	}
	for _, key := range []string{"merge-once", "merge-retry-new-key"} {
		b, e = h.s.MergeCart(ctx, h.buyer, owner.GuestToken, key)
		again := decodeCommerce[backoffice.CartView](t, b, e)
		if again.Items[0].Quantity != 5 {
			t.Fatal("merge doubled")
		}
	}
	if _, e = h.s.MergeCart(ctx, h.other, owner.GuestToken, "foreign-merge"); e != backoffice.ErrForbidden {
		t.Fatal("guest cart ownership", e)
	}
	if _, e = h.s.ReadCart(ctx, owner); e != backoffice.ErrForbidden {
		t.Fatal("consumed guest token still reads")
	}
	current := h.cart(t, h.buyer)
	if _, e = h.s.MutateCart(ctx, backoffice.CartOwner{Actor: h.buyer}, backoffice.CartMutation{Action: "add", OfferID: "offer1", Quantity: 1, Version: current.Version}, h.key()); e != backoffice.ErrStock {
		t.Fatal("oversized increment accepted", e)
	}
	if h.cart(t, h.other).ItemCount != 0 {
		t.Fatal("foreign cart leak")
	}
}
func TestCommerceAddressOwnerAndVersion(t *testing.T) {
	h := newCommerceHarness(t)
	a := h.address(t, h.buyer)
	if _, e := h.s.SaveAddress(context.Background(), h.other, backoffice.AddressInput{ID: a.ID, Version: a.Version, Delete: true}, h.key()); e != backoffice.ErrNotFound {
		t.Fatal(e)
	}
	if _, e := h.s.SaveAddress(context.Background(), h.buyer, backoffice.AddressInput{ID: a.ID, Version: 0, Delete: true}, h.key()); e != backoffice.ErrConflict {
		t.Fatal(e)
	}
	addresses, e := h.s.ReadAddresses(context.Background(), h.other)
	if e != nil || len(addresses) != 0 {
		t.Fatal("foreign addresses")
	}
}
func TestCommercePreviewHasNoReservationAndPurchaseHasTwoGroups(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	q := h.quote(t, h.buyer, "local")
	w, _ := h.repo.Snapshot(context.Background())
	if w.Offers[0].Reserved != 0 || len(w.Orders) != 0 || q.TotalToman != "260000" {
		t.Fatal("preview reserved or totals wrong", q.TotalToman)
	}
	if _, e := h.s.AcceptQuote(context.Background(), h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: "1"}, h.key()); e != backoffice.ErrCheckoutChanged {
		t.Fatal("client amount trusted", e)
	}
	p := h.accept(t, h.buyer, q)
	if len(p.OrderIDs) != 2 || p.PaymentStatus != "unpaid" || p.OrderStatus != "placed" || len(p.Events) != 1 {
		t.Fatal("order status/groups")
	}
	w, _ = h.repo.Snapshot(context.Background())
	if w.Offers[0].Reserved != 1 || w.Offers[1].Reserved != 1 || len(w.Orders) != 2 {
		t.Fatal("reservation/group mismatch")
	}
	if h.cart(t, h.buyer).ItemCount != 0 {
		t.Fatal("placed cart not cleared")
	}
	b, e := h.s.AcceptQuote(context.Background(), h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key())
	again := decodeCommerce[backoffice.PurchaseView](t, b, e)
	if again.ID != p.ID {
		t.Fatal("purchase duplicated")
	}
}
func TestCommerceFundingFirstThenAtomicPaymentAndExcess(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 1)
	p := h.accept(t, h.buyer, h.quote(t, h.buyer, "card-transfer"))
	d := h.deposit(t, h.buyer, p.ID, "card-transfer", "150000", "DEP-100")
	if h.wallet(t, h.buyer).BalanceToman != "0" {
		t.Fatal("declared deposit credited")
	}
	h.approve(t, d)
	if h.wallet(t, h.buyer).BalanceToman != "150000" {
		t.Fatal("credit amount")
	}
	ps, e := h.s.ReadPurchases(context.Background(), h.buyer, p.ID)
	if e != nil || ps[0].PaymentStatus != "unpaid" || len(ps[0].Events) != 2 {
		t.Fatal("credit paid order")
	}
	p = ps[0]
	in := backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, e := h.s.ConfirmPurchase(context.Background(), h.finance, in, fmt.Sprintf("confirm-%d", i))
			errs <- e
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	wallet := h.wallet(t, h.buyer)
	if wallet.BalanceToman != "30000" || len(wallet.Entries) != 2 {
		t.Fatal("duplicate debit", wallet.BalanceToman, len(wallet.Entries))
	}
	ps, _ = h.s.ReadPurchases(context.Background(), h.buyer, p.ID)
	if ps[0].PaymentStatus != "paid" || ps[0].OrderStatus != "placed" || len(ps[0].Events) != 3 {
		t.Fatal("paid status/history")
	}
	w, _ := h.repo.Snapshot(context.Background())
	if w.Offers[0].Stock != 4 || w.Offers[0].Reserved != 0 || w.Orders[0].Payment != "paid" {
		t.Fatal("stock/order atomicity")
	}
	raw, _ := json.Marshal(w)
	if !strings.Contains(string(raw), "deposit-credit") {
		t.Fatal("funding provenance missing")
	}
}
func TestCommerceInsufficientWalletDoesNotDebitOrConsumeStock(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 1)
	p := h.accept(t, h.buyer, h.quote(t, h.buyer, "local"))
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", "50000", "DEP-LOW"))
	if _, e := h.s.ConfirmPurchase(context.Background(), h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key()); e != backoffice.ErrInsufficientWallet {
		t.Fatal(e)
	}
	w, _ := h.repo.Snapshot(context.Background())
	if w.Offers[0].Stock != 5 || w.Offers[0].Reserved != 1 || len(w.Commerce.Entries) != 1 {
		t.Fatal("failed payment mutated state")
	}
}
func TestCommerceMethodCategoryDescendantsAndWholeOrderLimits(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "gold-offer", 1)
	cart := h.cart(t, h.buyer)
	a := h.address(t, h.buyer)
	in := backoffice.QuoteInput{CartVersion: cart.Version, AddressID: a.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "local"}
	if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); e != backoffice.ErrInvalidMethod {
		t.Fatal("gold branch local allowed", e)
	}
	h2 := newCommerceHarness(t)
	h2.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Price = 149900000; return nil })
	h2.add(t, h2.buyer, "offer1", 1)
	cart = h2.cart(t, h2.buyer)
	a = h2.address(t, h2.buyer)
	in = backoffice.QuoteInput{CartVersion: cart.Version, AddressID: a.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "card-transfer"}
	if _, e := h2.s.CreateQuote(context.Background(), h2.buyer, in, h2.key()); e != backoffice.ErrInvalidMethod {
		t.Fatal("shipping excluded from method cap", e)
	}
	if _, e := h2.s.CreateDeposit(context.Background(), h2.buyer, backoffice.DepositInput{MethodID: "card-transfer", AmountToman: "15000000.1", Date: "2026-11-22", Reference: "TOO-LARGE"}, h2.key()); e != backoffice.ErrInvalidMethod {
		t.Fatal("standalone funding cap", e)
	}
}
func TestCommerceValidSessionFreezesPriceAndExpiredNeedsCustomerConsent(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 1)
	p := h.accept(t, h.buyer, h.quote(t, h.buyer, "local"))
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Price = 2000000; return nil })
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", p.Quote.TotalToman, "FROZEN"))
	b, e := h.s.ConfirmPurchase(context.Background(), h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	paid := decodeCommerce[backoffice.PurchaseView](t, b, e)
	if paid.PaymentStatus != "paid" || paid.Quote.TotalToman != "120000" {
		t.Fatal("accepted price not frozen")
	}
	h2 := newCommerceHarness(t)
	h2.add(t, h2.buyer, "offer1", 1)
	p = h2.accept(t, h2.buyer, h2.quote(t, h2.buyer, "local"))
	h2.clock.Advance(31 * time.Minute)
	h2.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Price = 2000000; return nil })
	h2.approve(t, h2.deposit(t, h2.buyer, p.ID, "local", "250000", "LATE"))
	rows, e := h2.s.ReadPurchases(context.Background(), h2.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	p = rows[0]
	b, e = h2.s.ConfirmPurchase(context.Background(), h2.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h2.key())
	waiting := decodeCommerce[backoffice.PurchaseView](t, b, e)
	if waiting.Status != "requires-reconfirmation" || waiting.PaymentStatus != "unpaid" || h2.wallet(t, h2.buyer).BalanceToman != "250000" {
		t.Fatal("late money completed expired order")
	}
	w, _ := h2.repo.Snapshot(context.Background())
	if w.Offers[0].Reserved != 0 {
		t.Fatal("expired reservation retained")
	}
	input := backoffice.QuoteInput{PurchaseID: p.ID, AddressID: p.Quote.Address.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "local"}
	b, e = h2.s.CreateQuote(context.Background(), h2.buyer, input, h2.key())
	fresh := decodeCommerce[backoffice.QuoteView](t, b, e)
	if fresh.TotalToman != "220000" {
		t.Fatal("requote price")
	}
	if _, e = h2.s.AcceptQuote(context.Background(), h2.buyer, backoffice.AcceptQuoteInput{QuoteID: fresh.ID, AcceptedTotalToman: "120000"}, h2.key()); e != backoffice.ErrCheckoutChanged {
		t.Fatal("stale consent accepted")
	}
	reaccepted := h2.accept(t, h2.buyer, fresh)
	if reaccepted.ID != p.ID || reaccepted.PaymentStatus != "unpaid" {
		t.Fatal("requote created duplicate or paid automatically")
	}
	b, e = h2.s.ConfirmPurchase(context.Background(), h2.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: reaccepted.Version}, h2.key())
	paid = decodeCommerce[backoffice.PurchaseView](t, b, e)
	if paid.PaymentStatus != "paid" || h2.wallet(t, h2.buyer).BalanceToman != "30000" {
		t.Fatal("reconfirmed payment")
	}
}
func TestCommerceConcurrentLastStockAndRefundHolds(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 5)
	h.add(t, h.other, "offer1", 5)
	q1 := h.quote(t, h.buyer, "local")
	q2 := h.quote(t, h.other, "local")
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i, q := range []backoffice.QuoteView{q1, q2} {
		wg.Add(1)
		a := h.buyer
		if i == 1 {
			a = h.other
		}
		go func(a identity.User, q backoffice.QuoteView, i int) {
			defer wg.Done()
			_, e := h.s.AcceptQuote(context.Background(), a, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, fmt.Sprintf("last-stock-%d", i))
			results <- e
		}(a, q, i)
	}
	wg.Wait()
	close(results)
	ok, fail := 0, 0
	for e := range results {
		if e == nil {
			ok++
		} else if e == backoffice.ErrStock {
			fail++
		} else {
			t.Fatal(e)
		}
	}
	if ok != 1 || fail != 1 {
		t.Fatal("oversell", ok, fail)
	}
	h2 := newCommerceHarness(t)
	h2.approve(t, h2.deposit(t, h2.buyer, "", "card-transfer", "1000000", "REFUND-FUND"))
	b, e := h2.s.RequestRefund(context.Background(), h2.buyer, backoffice.RefundInput{AmountToman: "900000", Instructions: "اطلاعات ساختگی پرداخت دستی"}, h2.key())
	r := decodeCommerce[backoffice.RefundView](t, b, e)
	if h2.wallet(t, h2.buyer).AvailableToman != "100000" {
		t.Fatal("refund not held")
	}
	h2.add(t, h2.buyer, "offer1", 1)
	p := h2.accept(t, h2.buyer, h2.quote(t, h2.buyer, "wallet"))
	if _, e = h2.s.ConfirmPurchase(context.Background(), h2.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h2.key()); e != backoffice.ErrInsufficientWallet {
		t.Fatal("refund hold spent")
	}
	in := backoffice.RefundResolveInput{ID: r.ID, Version: r.Version, Complete: true, Reference: "REFUND-001", Date: "2026-11-22", Reason: "بازپرداخت دستی انجام شد"}
	for _, key := range []string{"refund-once", "refund-retry"} {
		b, e = h2.s.ResolveRefund(context.Background(), h2.finance, in, key)
		done := decodeCommerce[backoffice.RefundView](t, b, e)
		if done.Status != "completed" {
			t.Fatal("refund completion")
		}
	}
	wallet := h2.wallet(t, h2.buyer)
	if wallet.BalanceToman != "100000" || wallet.HeldToman != "0" || len(wallet.Entries) != 2 {
		t.Fatal("refund double debit")
	}
}
func TestCommerceReviewPermissionsAndVerifiedReferenceUniqueness(t *testing.T) {
	h := newCommerceHarness(t)
	d := h.deposit(t, h.buyer, "", "card-transfer", "100000", "CUSTOMER-REF-A")
	in := backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedAmountToman: d.AmountToman, VerifiedDate: d.Date, VerifiedReference: "BANK-ACTUAL"}
	if _, e := h.s.ReviewDeposit(context.Background(), h.ops, in, "ops-approval"); e != backoffice.ErrForbidden {
		t.Fatal("operations approved money", e)
	}
	b, e := h.s.ReviewDeposit(context.Background(), h.finance, in, "finance-approval")
	decodeCommerce[backoffice.DepositView](t, b, e)
	other := h.deposit(t, h.other, "", "card-transfer", "100000", "CUSTOMER-REF-B")
	in.ID = other.ID
	in.Version = other.Version
	if _, e = h.s.ReviewDeposit(context.Background(), h.finance, in, "duplicate-bank-reference"); e != backoffice.ErrConflict {
		t.Fatal("same actual bank transfer credited twice", e)
	}
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == h.finance.ID {
				w.Users[i].Role = identity.RoleViewer
			}
		}
		return nil
	})
	in.ID = d.ID
	in.Version = d.Version
	if _, e = h.s.ReviewDeposit(context.Background(), h.finance, in, "finance-approval"); e != backoffice.ErrForbidden {
		t.Fatal("revoked permission bypassed via replay", e)
	}
	v, e := h.s.Read(context.Background(), h.admin)
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(v)
	if strings.Contains(string(raw), `"commerce"`) || strings.Contains(string(raw), "BANK-ACTUAL") {
		t.Fatal("private commerce leaked to generic workspace")
	}
}

func TestCommercePaidCancellationReturnsWalletOnce(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "card-transfer", q.TotalToman, "CANCEL-001")
	h.approve(t, d)
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	b, e = h.s.CancelPurchase(ctx, h.buyer, p.ID, "لغو ساختگی مشتری", h.key())
	requested := decodeCommerce[backoffice.PurchaseView](t, b, e)
	if !requested.CancellationRequested || requested.PaymentStatus != "paid" {
		t.Fatal("customer bypassed staff cancellation")
	}
	b, e = h.s.AdminCancelPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: requested.Version, Reason: "لغو پیش از تأمین"}, h.key())
	cancelled := decodeCommerce[backoffice.PurchaseView](t, b, e)
	if cancelled.PaymentStatus != "returned-to-wallet" || cancelled.OrderStatus != "cancelled" || cancelled.WalletBalanceToman != q.TotalToman {
		t.Fatal(cancelled)
	}
	b, e = h.s.AdminCancelPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: requested.Version, Reason: "لغو پیش از تأمین"}, h.key())
	decodeCommerce[backoffice.PurchaseView](t, b, e)
	state, _ := h.repo.Snapshot(ctx)
	if len(state.Commerce.Entries) != 3 || state.Offers[0].Stock != 5 {
		t.Fatal("duplicate cancellation credit/restock")
	}
}
func TestCommerceMethodRestrictionRecheckedWithoutRepricingFrozenOrder(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "METHOD-001"))
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Commerce.Settings.Methods {
			if w.Commerce.Settings.Methods[i].ID == "local" {
				w.Commerce.Settings.Methods[i].Max = 1
			}
		}
		return nil
	})
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	invalid := decodeCommerce[backoffice.PurchaseView](t, b, e)
	if invalid.Status != "requires-reconfirmation" || invalid.PaymentStatus != "unpaid" || invalid.WalletBalanceToman != q.TotalToman {
		t.Fatal(invalid)
	}
	state, _ := h.repo.Snapshot(ctx)
	if state.Offers[0].Reserved != 0 || state.Offers[0].Stock != 5 || len(state.Commerce.Entries) != 1 {
		t.Fatal("restriction consumed money/stock")
	}
}

func TestCommerceOrderReadWithoutFinanceHidesWalletAndDeposits(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	p := h.accept(t, h.buyer, q)
	h.deposit(t, h.buyer, p.ID, "card-transfer", q.TotalToman, "PRIVATE-001")
	v, e := h.s.ReadCommerceAdmin(ctx, h.ops)
	if e != nil {
		t.Fatal(e)
	}
	if len(v.Purchases) != 1 || len(v.Deposits) != 0 || len(v.Purchases[0].Deposits) != 0 || v.Purchases[0].WalletAvailableToman != "" || v.Purchases[0].ShortfallToman != "" {
		t.Fatal("finance metadata leaked to operations")
	}
}

func TestCommerceSellerOrdersRequireOwnStoreGrantAndHidePrivateData(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	p := h.accept(t, h.buyer, h.quote(t, h.buyer, "card-transfer"))
	var seller identity.User
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "checkout-reader", Name: "آزمون", Scope: identity.ScopeStore, Permissions: []string{"orders.read"}, Version: 1})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "seller-test", RoleID: "checkout-reader", StoreID: "store1"})
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: "seller-test", StoreID: "store1"})
		for _, u := range w.Users {
			if u.ID == "seller-test" {
				seller = identity.User{ID: u.ID, Active: u.Active, AccessVersion: u.AccessVersion, Surface: identity.SurfaceSeller}
			}
		}
		return nil
	})
	rows, e := h.s.ReadSellerCheckoutOrders(ctx, seller, "")
	if e != nil || len(rows) != 1 || rows[0].StoreID != "store1" {
		t.Fatal(rows, e)
	}
	data, _ := json.Marshal(rows)
	for _, private := range []string{"recipient", "phone", "street", "receipt", "wallet"} {
		if strings.Contains(string(data), `"`+private+`":`) {
			t.Fatal("seller privacy", private)
		}
	}
	if _, e = h.s.ReadSellerCheckoutOrders(ctx, seller, p.ID+":store2"); !errors.Is(e, backoffice.ErrNotFound) {
		t.Fatal("foreign store order", e)
	}
}

func TestCommerceShippingAndAddressSnapshotsSurviveSettingsEdits(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	p := h.accept(t, h.buyer, q)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].Fee = 400000
		w.Stores[0].LeadDays = 4
		for i := range w.Commerce.Addresses {
			if w.Commerce.Addresses[i].ID == q.Address.ID {
				w.Commerce.Addresses[i].Street = "نشانی تازه ساختگی"
				w.Commerce.Addresses[i].Version++
			}
		}
		return nil
	})
	h.add(t, h.buyer, "offer1", 1)
	fresh := h.quote(t, h.buyer, "card-transfer")
	if fresh.Shipments[0].FeeToman != "40000" || fresh.Shipments[0].LeadDays != 4 {
		t.Fatal("new shipping settings not consumed", fresh)
	}
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	old := rows[0].Quote
	if old.ShippingToman != q.ShippingToman || old.Shipments[0].LeadDays != 1 || old.Address.Street != q.Address.Street {
		t.Fatal("accepted shipping/address snapshot changed")
	}
}

func TestCommercePublicNumbersConcurrentPrivateAndRetry(t *testing.T) {
	h := newCommerceHarness(t)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.other, "offer1", 1)
	q1, q2 := h.quote(t, h.buyer, "local"), h.quote(t, h.other, "local")
	results := make(chan backoffice.PurchaseView, 2)
	var wg sync.WaitGroup
	for i, q := range []backoffice.QuoteView{q1, q2} {
		a := h.buyer
		if i == 1 {
			a = h.other
		}
		wg.Add(1)
		go func(i int, a identity.User, q backoffice.QuoteView) {
			defer wg.Done()
			raw, err := h.s.AcceptQuote(context.Background(), a, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, fmt.Sprintf("number-%d", i))
			if err != nil {
				t.Error(err)
				return
			}
			var p backoffice.PurchaseView
			if err = json.Unmarshal(raw, &p); err != nil {
				t.Error(err)
				return
			}
			results <- p
		}(i, a, q)
	}
	wg.Wait()
	close(results)
	numbers := map[string]bool{}
	for p := range results {
		numbers[p.Number] = true
	}
	if len(numbers) != 2 || !numbers["1000"] || !numbers["1001"] {
		t.Fatal(numbers)
	}
	mine, err := h.s.ReadPurchases(context.Background(), h.buyer, "")
	if err != nil || len(mine) != 1 {
		t.Fatal(err)
	}
	own := mine[0]
	byNumber, err := h.s.ReadPurchases(context.Background(), h.buyer, own.Number)
	if err != nil || len(byNumber) != 1 || byNumber[0].ID != own.ID {
		t.Fatal("number lookup", err)
	}
	if _, err = h.s.ReadPurchases(context.Background(), h.other, own.Number); err != backoffice.ErrNotFound {
		t.Fatal("enumerable number bypassed ownership", err)
	}
	raw, err := h.s.AcceptQuote(context.Background(), h.buyer, backoffice.AcceptQuoteInput{QuoteID: q1.ID, AcceptedTotalToman: q1.TotalToman}, "number-0")
	retry := decodeCommerce[backoffice.PurchaseView](t, raw, err)
	if retry.ID != own.ID || retry.Number != own.Number {
		t.Fatal("retry assigned new order number")
	}
}
