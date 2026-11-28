package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

func storePolicyHarness(t *testing.T) *commerceHarness {
	h := newCommerceHarness(t)
	if e := h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.Sales = &commerce.SalesSettings{Version: 1, Methods: []commerce.SalesMethod{{ID: "cash", Name: "نقدی", Active: true, PaymentMethodIDs: []string{"card-transfer"}}, {ID: "credit", Name: "اعتباری آزمون", Active: true, MarkupBPS: 400, PaymentMethodIDs: []string{"local"}}}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	return h
}
func policyInput() backoffice.StoreEligibilityInput {
	return backoffice.StoreEligibilityInput{CategoryMode: "all", Methods: []store.SalePermission{{MethodID: "cash", DisplayName: "نام نقدی ساختگی"}, {MethodID: "credit", DisplayName: "نام اعتباری ساختگی"}}}
}
func savePolicy(t *testing.T, h *commerceHarness, id string, in backoffice.StoreEligibilityInput) backoffice.AdminStoreRow {
	t.Helper()
	r, e := h.s.SaveStoreEligibility(context.Background(), h.admin, id, in, h.key())
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestStoreEligibilityTreeDenyPriorityLegacyAndNewDescendant(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	before, _ := h.repo.Snapshot(ctx)
	in := policyInput()
	in.CategoryMode = "selected"
	in.AllowedCategories = []string{"regular", "coin"}
	in.DeniedCategories = []string{"gold"}
	r := savePolicy(t, h, "store1", in)
	w, _ := h.repo.Snapshot(ctx)
	for _, c := range []struct {
		id    string
		allow bool
	}{{"regular", true}, {"gold", false}, {"coin", false}, {"unknown", false}} {
		if got := backoffice.StoreAllowsCategory(w, r.Store, c.id); got != c.allow {
			t.Fatal(c.id, got)
		}
	}
	w.Categories = append(w.Categories, catalog.Category{ID: "new-child", ParentID: "coin"}, catalog.Category{ID: "new-digital", ParentID: "regular"}, catalog.Category{ID: "cycle-a", ParentID: "cycle-b"}, catalog.Category{ID: "cycle-b", ParentID: "cycle-a"})
	if backoffice.StoreAllowsCategory(w, r.Store, "new-child") || !backoffice.StoreAllowsCategory(w, r.Store, "new-digital") {
		t.Fatal("descendant rules")
	}
	all := r.Store
	all.Eligibility = &store.Eligibility{CategoryMode: "all"}
	if backoffice.StoreAllowsCategory(w, all, "cycle-a") {
		t.Fatal("cycle")
	}
	if !backoffice.StoreAllowsMethod(before.Stores[0], "credit") || !backoffice.StoreAllowsCategory(before, before.Stores[0], "coin") {
		t.Fatal("legacy permissions changed")
	}
	if !reflect.DeepEqual(before.Users, w.Users) || !reflect.DeepEqual(before.Orders, w.Orders) || !reflect.DeepEqual(before.Ledger, w.Ledger) || !reflect.DeepEqual(before.Offers, w.Offers) || !reflect.DeepEqual(before.Commerce, w.Commerce) {
		t.Fatal("policy edited protected data")
	}
}
func TestStoreEligibilityAliasOnlyAfterSelectionAndHistoricalSnapshot(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	savePolicy(t, h, "store1", policyInput())
	h.add(t, h.buyer, "offer1", 1)
	address := h.address(t, h.buyer)
	cart := h.cart(t, h.buyer)
	input := backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"store1": "post"}}
	v, e := h.s.PreviewCheckout(ctx, h.buyer, input)
	if e != nil || v.Quote.Shipments[0].StoreName != "فروشگاه یک" {
		t.Fatal("alias before selecting method", e)
	}
	input.PaymentMethodID = "card-transfer"
	v, e = h.s.PreviewCheckout(ctx, h.buyer, input)
	if e != nil || v.Quote.Shipments[0].StoreName != "نام نقدی ساختگی" {
		t.Fatal("cash alias", e)
	}
	q := h.quote(t, h.buyer, "local")
	if q.Shipments[0].StoreName != "نام اعتباری ساختگی" || q.TotalToman != "124000" {
		t.Fatal("credit alias or markup", q.TotalToman)
	}
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "policy-funding")
	h.approve(t, d)
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	if p.OrderStatus != "placed" {
		t.Fatal("confirmation", p.OrderStatus)
	}
	before, _ := h.repo.Snapshot(ctx)
	if before.Orders[0].StoreDisplayName != "نام اعتباری ساختگی" || before.Orders[0].StoreID != "store1" || before.Commerce.Quotes[0].Lines[0].StorePolicyVersion != 1 {
		t.Fatal("stable identity or alias snapshot")
	}
	in := policyInput()
	in.Version = 1
	in.Methods[1].DisplayName = "نام جدید اعتباری"
	savePolicy(t, h, "store1", in)
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Transfers, after.Transfers) || after.Stores[0].Name != "فروشگاه یک" {
		t.Fatal("historical/financial identity rewritten")
	}
}
func TestStoreEligibilityChangeInvalidatesReservedPurchaseWithoutConsumingFunds(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	savePolicy(t, h, "store1", policyInput())
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "policy-late-funding"))
	balance := h.wallet(t, h.buyer).BalanceToman
	in := policyInput()
	in.Version = 1
	in.Methods = in.Methods[:1]
	savePolicy(t, h, "store1", in)
	before, _ := h.repo.Snapshot(ctx)
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	result := decodeCommerce[backoffice.PurchaseView](t, b, e)
	after, _ := h.repo.Snapshot(ctx)
	if result.Status != "requires-reconfirmation" || h.wallet(t, h.buyer).BalanceToman != balance || len(after.Commerce.Entries) != len(before.Commerce.Entries) || after.Offers[0].Reserved != 0 {
		t.Fatal("policy bypass or financial consumption", result.OrderStatus)
	}
	if after.Commerce.Quotes[0].Lines[0].UnitPrice != before.Commerce.Quotes[0].Lines[0].UnitPrice || after.Commerce.Quotes[0].Shipments[0].StoreName != before.Commerce.Quotes[0].Shipments[0].StoreName {
		t.Fatal("terms rewritten")
	}
}
func TestStoreEligibilityWalletOnlyAlsoRequiresPolicyReconfirmation(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	savePolicy(t, h, "store1", policyInput())
	h.add(t, h.buyer, "offer1", 1)
	address := h.address(t, h.buyer)
	_, e := h.s.CreditCustomerWallet(ctx, h.admin, backoffice.AdminCreditInput{UserID: h.buyer.ID, AmountToman: "120000", Title: "اعتبار ساختگی", Description: "آزمون تغییر سیاست فروشگاه"}, h.key())
	if e != nil {
		t.Fatal(e)
	}
	in := backoffice.QuoteInput{ContractVersion: 2, CartVersion: h.cart(t, h.buyer).Version, AddressID: address.ID, Shipping: map[string]string{"store1": "post"}, WalletToman: "120000"}
	b, e := h.s.CreateQuote(ctx, h.buyer, in, h.key())
	q := decodeCommerce[backoffice.QuoteView](t, b, e)
	changed := policyInput()
	changed.Version = 1
	changed.Methods[0].DisplayName = "نام نقدی تازه"
	savePolicy(t, h, "store1", changed)
	_, e = h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, h.key())
	wallet := h.wallet(t, h.buyer)
	if e != backoffice.ErrCheckoutChanged || wallet.BalanceToman != "120000" || wallet.HeldToman != "0" {
		t.Fatal("wallet-only policy bypass", e, wallet)
	}
}
func TestStoreEligibilityMixedCartMethodsAndDirectOfferGuards(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	in := policyInput()
	in.Methods = in.Methods[:1]
	savePolicy(t, h, "store2", in)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	a := h.address(t, h.buyer)
	input := backoffice.QuoteInput{CartVersion: h.cart(t, h.buyer).Version, AddressID: a.ID, Shipping: map[string]string{"store1": "post", "store2": "post"}}
	v, e := h.s.PreviewCheckout(ctx, h.buyer, input)
	if e != nil {
		t.Fatal(e)
	}
	for _, m := range v.Methods {
		if m.ID == "local" && m.Eligible {
			t.Fatal("noncommon sales method allowed")
		}
	}
	input.PaymentMethodID = "local"
	if _, e = h.s.CreateQuote(ctx, h.buyer, input, h.key()); !errors.Is(e, backoffice.ErrInvalidMethod) {
		t.Fatal("direct quote bypass", e)
	}
	denied := policyInput()
	denied.DeniedCategories = []string{"regular"}
	savePolicy(t, h, "store1", denied)
	w, _ := h.repo.Snapshot(ctx)
	d, e := backoffice.BuildProductDetail(w, "phone", "phone-v", "offer1")
	if e != nil {
		t.Fatal(e)
	}
	for _, o := range d.Offers {
		if o.ID == "offer1" && (!o.Restricted || o.Available || o.Price != nil) {
			t.Fatal("restricted offer still purchasable")
		}
	}
	price := int64(1)
	if e = h.s.Execute(ctx, h.ops, backoffice.Command{Action: "offer.price.update", Target: "offer1", Version: 1, Offer: &backoffice.OfferInput{Price: &price}, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("offer update bypass", e)
	}

	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: h.ops.ID, StoreID: "store1"})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: h.ops.ID, StoreID: "store1", RoleID: "store-stock"})
		return nil
	})
	seller := h.ops
	seller.Surface = identity.SurfaceSeller
	stock := 1
	if e = h.s.Execute(ctx, seller, backoffice.Command{Action: "offer.stock.update", Target: "offer1", Version: 1, Offer: &backoffice.OfferInput{Stock: &stock}, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("seller category policy bypass", e)
	}
	cart := h.cart(t, h.buyer)
	if cart.CanCheckout || !cart.Items[0].Restricted {
		t.Fatal("old cart still buyable")
	}
	if _, e = h.s.MutateCart(ctx, backoffice.CartOwner{Actor: h.other}, backoffice.CartMutation{Action: "add", OfferID: "offer1", Version: 1, Quantity: 1}, h.key()); e == nil {
		t.Fatal("direct add bypass")
	}
}
func TestStoreEligibilityValidationVersionsRetryAndAuthorization(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	in := policyInput()
	r, e := h.s.SaveStoreEligibility(ctx, h.admin, "store1", in, "policy-once")
	if e != nil {
		t.Fatal(e)
	}
	again, e := h.s.SaveStoreEligibility(ctx, h.admin, "store1", in, "policy-once")
	if e != nil || !reflect.DeepEqual(r, again) {
		t.Fatal("policy retry", e)
	}
	in.Methods[0].DisplayName = "متفاوت"
	if _, e = h.s.SaveStoreEligibility(ctx, h.admin, "store1", in, "policy-once"); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("key conflict")
	}
	if _, e = h.s.SaveStoreEligibility(ctx, h.admin, "store1", in, h.key()); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("stale version")
	}
	for _, actor := range []identity.User{h.finance, {ID: h.ops.ID, Surface: identity.SurfaceSeller, Active: true}} {
		if _, e = h.s.SaveStoreEligibility(ctx, actor, "store1", in, h.key()); !errors.Is(e, backoffice.ErrForbidden) {
			t.Fatal("policy permission", e)
		}
	}
	for _, bad := range []backoffice.StoreEligibilityInput{{Version: 1, CategoryMode: "invalid"}, {Version: 1, CategoryMode: "all", AllowedCategories: []string{"missing"}}, {Version: 1, CategoryMode: "all", DeniedCategories: []string{"gold", "gold"}}, {Version: 1, CategoryMode: "all", Methods: []store.SalePermission{{MethodID: "unknown"}}}} {
		if _, e = h.s.SaveStoreEligibility(ctx, h.admin, "store1", bad, h.key()); !errors.Is(e, backoffice.ErrInvalid) {
			t.Fatal("invalid policy", e)
		}
	}
}

func TestStoreEligibilityConcurrentWritesAndRevokedRole(t *testing.T) {
	for _, same := range []bool{false, true} {
		t.Run(fmt.Sprint(same), func(t *testing.T) {
			h := storePolicyHarness(t)
			ctx := context.Background()
			var wg sync.WaitGroup
			var success atomic.Int32
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					key := "policy-concurrent"
					if !same {
						key = fmt.Sprint(key, i)
					}
					_, e := h.s.SaveStoreEligibility(ctx, h.admin, "store1", policyInput(), key)
					if e == nil {
						success.Add(1)
					} else if !errors.Is(e, backoffice.ErrConflict) {
						t.Error(e)
					}
				}(i)
			}
			wg.Wait()
			want := int32(1)
			if same {
				want = 10
			}
			if success.Load() != want {
				t.Fatal("concurrent policy", success.Load())
			}
			w, _ := h.repo.Snapshot(ctx)
			if w.Stores[0].Version != 1 || w.Stores[0].Eligibility.Version != 1 {
				t.Fatal("double policy mutation")
			}
			h.repo.Update(ctx, func(w *backoffice.Workspace) error {
				for i := range w.Users {
					if w.Users[i].ID == h.admin.ID {
						w.Users[i].Role = identity.RoleViewer
					}
				}
				return nil
			})
			in := policyInput()
			in.Version = 1
			if _, e := h.s.SaveStoreEligibility(ctx, h.admin, "store1", in, "revoked-policy"); !errors.Is(e, backoffice.ErrForbidden) {
				t.Fatal("revoked policy permission", e)
			}
		})
	}
}

func TestStoreEligibilityStalePreviewCannotReserve(t *testing.T) {
	h := storePolicyHarness(t)
	ctx := context.Background()
	savePolicy(t, h, "store1", policyInput())
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	in := policyInput()
	in.Version = 1
	in.Methods[0].DisplayName = "نام نقدی تازه"
	savePolicy(t, h, "store1", in)
	if _, e := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key()); !errors.Is(e, backoffice.ErrCheckoutChanged) {
		t.Fatal("stale policy preview reserved", e)
	}
	w, _ := h.repo.Snapshot(ctx)
	if w.Offers[0].Reserved != 0 || len(w.Commerce.Purchases) > 0 || len(w.Commerce.Entries) > 0 {
		t.Fatal("rejected preview moved inventory/money")
	}
}
