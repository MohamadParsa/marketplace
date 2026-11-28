package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func axisHarness(t *testing.T) *commerceHarness {
	h := newCommerceHarness(t)
	e := h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.Attributes = []catalog.AttributeDefinition{{ID: "old", Code: "old", Name: "رنگ", Kind: "enum", Options: []string{"سیاه", "سفید"}, Version: 1}, {ID: "new", Code: "new", Name: "پوشش", Kind: "enum", Options: []string{"مات", "براق"}, Version: 1}}
		w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "old", Axis: true, Filter: true, Card: true, Position: 0}}
		w.Variants[0].Values = []catalog.AttributeValue{{AttributeID: "old", Value: "سیاه"}}
		w.Variants = append(w.Variants, catalog.Variant{ID: "phone-v2", ProductID: "phone", SKU: "old-sku", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "old", Value: "سفید"}}})
		w.Offers = append(w.Offers, catalog.Offer{ID: "offer3", ProductID: "phone", VariantID: "phone-v2", StoreID: "store1", Price: 1100000, Stock: 5, Version: 1})
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return h
}
func axisInput() backoffice.AxisConversionInput {
	return backoffice.AxisConversionInput{CategoryID: "regular", Version: 1, AxisIDs: []string{"new"}, Mapping: []backoffice.AxisMapping{{VariantID: "phone-v", Values: []catalog.AttributeValue{{AttributeID: "new", Value: "مات"}}}, {VariantID: "phone-v2", Values: []catalog.AttributeValue{{AttributeID: "new", Value: "براق"}}}}}
}
func axisPlan(t *testing.T, h *commerceHarness) backoffice.AxisPlanView {
	t.Helper()
	v, e := h.s.PreviewAxisConversion(context.Background(), h.admin, axisInput())
	if e != nil || !v.Ready {
		t.Fatal("preview", v, e)
	}
	p, e := h.s.PrepareAxisConversion(context.Background(), h.admin, axisInput(), v.Fingerprint, h.key())
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestAxisPreviewIsReadOnlyAndRejectsCollisionsAndMissingMapping(t *testing.T) {
	h := axisHarness(t)
	ctx := context.Background()
	before, _ := h.repo.Snapshot(ctx)
	in := axisInput()
	v, e := h.s.PreviewAxisConversion(ctx, h.admin, in)
	if e != nil || !v.Ready || v.Impact.Variants != 2 || v.Impact.Offers != 3 {
		t.Fatal(v, e)
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("preview wrote data")
	}
	in.Mapping[1].Values[0].Value = "مات"
	v, e = h.s.PreviewAxisConversion(ctx, h.admin, in)
	if e != nil || v.Ready || v.Conflicts[0].Code != "combination_collision" {
		t.Fatal("collision", v, e)
	}
	if _, e = h.s.PrepareAxisConversion(ctx, h.admin, in, v.Fingerprint, "collision"); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal(e)
	}
	in = axisInput()
	in.Mapping = in.Mapping[:1]
	v, e = h.s.PreviewAxisConversion(ctx, h.admin, in)
	if e != nil || v.Ready {
		t.Fatal("missing mapping")
	}
	in = axisInput()
	in.Mapping[0].Values = nil
	v, e = h.s.PreviewAxisConversion(ctx, h.admin, in)
	if e != nil || v.Ready {
		t.Fatal("missing value")
	}
	in = axisInput()
	in.Mapping[0].VariantID = "foreign"
	if _, e = h.s.PreviewAxisConversion(ctx, h.admin, in); !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal("foreign variant", e)
	}
	if _, e = h.s.ReadAxisSource(ctx, h.ops, "regular"); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("dedicated grant required", e)
	}
}
func TestAxisApplyAtomicRetryHistoryAndCustomerReview(t *testing.T) {
	h := axisHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Quotes = append(w.Commerce.Quotes, commerce.Quote{ID: "unaccepted", Status: "preview", Lines: []commerce.LineSnapshot{{ProductID: "phone", VariantLabel: "سیاه"}}}, commerce.Quote{ID: "historical", Status: "confirmed", PurchaseID: "past", Lines: []commerce.LineSnapshot{{ProductID: "phone", VariantLabel: "سیاه", UnitPrice: 12345}}})
		return nil
	})
	before, _ := h.repo.Snapshot(ctx)
	preview, e := h.s.PreviewAxisConversion(ctx, h.admin, axisInput())
	if e != nil {
		t.Fatal(e)
	}
	plan, e := h.s.PrepareAxisConversion(ctx, h.admin, axisInput(), preview.Fingerprint, "prepare")
	if e != nil {
		t.Fatal(e)
	}
	again, e := h.s.PrepareAxisConversion(ctx, h.admin, axisInput(), preview.Fingerprint, "prepare")
	if e != nil || !reflect.DeepEqual(plan, again) {
		t.Fatal("prepare retry", e)
	}
	if _, e = h.s.ApplyAxisConversion(ctx, h.admin, plan.ID, plan.Fingerprint, false); !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal("explicit approval required")
	}
	var wg sync.WaitGroup
	receipts := make(chan backoffice.AxisReceipt, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := h.s.ApplyAxisConversion(ctx, h.admin, plan.ID, plan.Fingerprint, true)
			if e != nil {
				t.Error(e)
			}
			receipts <- r
		}()
	}
	wg.Wait()
	close(receipts)
	var first *backoffice.AxisReceipt
	for r := range receipts {
		if first == nil {
			c := r
			first = &c
		} else if !reflect.DeepEqual(*first, r) {
			t.Fatal("unstable retry")
		}
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Offers, after.Offers) || !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Users, after.Users) || !reflect.DeepEqual(before.Commerce.Settings, after.Commerce.Settings) || !reflect.DeepEqual(before.Commerce.Quotes[1], after.Commerce.Quotes[1]) {
		t.Fatal("history/stock/identity changed")
	}
	if after.Variants[0].Version != before.Variants[0].Version+1 || after.Variants[0].ID != before.Variants[0].ID || after.Variants[0].Values[0].Value != "مات" || len(after.AxisPlans[0].PreviousVariants) != 2 || after.Commerce.Quotes[0].Status != "expired" {
		t.Fatal("conversion or archive")
	}
	cart := h.cart(t, h.buyer)
	if cart.CanCheckout || !cart.Items[0].ReviewRequired || cart.Items[0].PreviousVariantLabel != "سیاه" || cart.Items[0].VariantLabel != "مات" {
		t.Fatal("missing customer consent", cart)
	}
	b, e := h.s.MutateCart(ctx, backoffice.CartOwner{Actor: h.buyer}, backoffice.CartMutation{Action: "set", OfferID: "offer1", Quantity: 2, Version: cart.Version}, h.key())
	cart = decodeCommerce[backoffice.CartView](t, b, e)
	if !cart.Items[0].ReviewRequired {
		t.Fatal("quantity bypassed review")
	}
	address := h.address(t, h.buyer)
	_, e = h.s.PreviewCheckout(ctx, h.buyer, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "card-transfer"})
	if !errors.Is(e, backoffice.ErrCheckoutChanged) {
		t.Fatal("checkout bypassed review", e)
	}
	stale := cart.Items[0].VariantVersion - 1
	_, e = h.s.MutateCart(ctx, backoffice.CartOwner{Actor: h.buyer}, backoffice.CartMutation{Action: "review", OfferID: "offer1", Version: cart.Version, ReviewVariantVersion: &stale}, h.key())
	if !errors.Is(e, backoffice.ErrCheckoutChanged) {
		t.Fatal("stale consent")
	}
	version := cart.Items[0].VariantVersion
	b, e = h.s.MutateCart(ctx, backoffice.CartOwner{Actor: h.buyer}, backoffice.CartMutation{Action: "review", OfferID: "offer1", Version: cart.Version, ReviewVariantVersion: &version}, h.key())
	cart = decodeCommerce[backoffice.CartView](t, b, e)
	if !cart.CanCheckout || cart.Items[0].ReviewRequired {
		t.Fatal("explicit review failed")
	}
	visible, e := h.s.Read(ctx, h.admin)
	if e != nil || visible.AxisPlans != nil {
		t.Fatal("private plans leaked")
	}
	detail, e := backoffice.BuildProductDetail(after, "phone", "phone-v", "")
	if e != nil || len(detail.Axes) != 1 || detail.Axes[0].ID != "new" {
		t.Fatal("PDP axes stale", e)
	}
}
func TestAxisPlanStaleReservationExpiryAndRevokedPermission(t *testing.T) {
	for _, what := range []string{"price", "reserved-offer", "reserved-quote", "cart", "definition", "expired", "revoked"} {
		t.Run(what, func(t *testing.T) {
			h := axisHarness(t)
			ctx := context.Background()
			p := axisPlan(t, h)
			if what == "expired" {
				h.clock.Advance(16 * time.Minute)
			} else {
				h.repo.Update(ctx, func(w *backoffice.Workspace) error {
					switch what {
					case "price":
						w.Offers[0].Price++
					case "reserved-offer":
						w.Offers[0].Reserved = 1
					case "reserved-quote":
						w.Commerce.Quotes = append(w.Commerce.Quotes, commerce.Quote{ID: "reserved", Status: "reserved", Lines: []commerce.LineSnapshot{{ProductID: "phone"}}})
					case "cart":
						w.Commerce.Carts = append(w.Commerce.Carts, commerce.Cart{ID: "new-cart", Lines: []commerce.CartLine{{OfferID: "offer1", Quantity: 1}}})
					case "definition":
						w.Attributes[1].Version++
					case "revoked":
						for i := range w.Users {
							if w.Users[i].ID == h.admin.ID {
								w.Users[i].Role = identity.RoleViewer
							}
						}
					}
					return nil
				})
			}
			before, _ := h.repo.Snapshot(ctx)
			_, e := h.s.ApplyAxisConversion(ctx, h.admin, p.ID, p.Fingerprint, true)
			if e == nil {
				t.Fatal("stale plan applied")
			}
			after, _ := h.repo.Snapshot(ctx)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("failed apply wrote data")
			}
		})
	}
	h := axisHarness(t)
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Reserved = 1; return nil })
	v, e := h.s.PreviewAxisConversion(context.Background(), h.admin, axisInput())
	if e != nil || v.Ready || v.Impact.ReservedOffers != 1 {
		t.Fatal("active reservation not reported")
	}
}

type failingAxisRepository struct{ backoffice.Repository }

func (r failingAxisRepository) Update(ctx context.Context, apply func(*backoffice.Workspace) error) error {
	return r.Repository.Update(ctx, func(w *backoffice.Workspace) error {
		if e := apply(w); e != nil {
			return e
		}
		return errors.New("isolated late write failure")
	})
}
func TestAxisLateWriteFailureRollsBackEntireConversion(t *testing.T) {
	h := axisHarness(t)
	ctx := context.Background()
	p := axisPlan(t, h)
	before, _ := h.repo.Snapshot(ctx)
	s := backoffice.NewService(failingAxisRepository{h.repo}, h.clock, &commerceIDs{})
	if _, e := s.ApplyAxisConversion(ctx, h.admin, p.ID, p.Fingerprint, true); e == nil {
		t.Fatal("failure ignored")
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("half conversion committed")
	}
	if _, e := h.s.ApplyAxisConversion(ctx, h.admin, p.ID, p.Fingerprint, true); e != nil {
		t.Fatal("rollback consumed plan", e)
	}
}

func TestAxisAdditionRemovalEmptyCategoryAndTechnicalBounds(t *testing.T) {
	for _, kind := range []string{"add", "remove-last", "empty", "over-limit"} {
		t.Run(kind, func(t *testing.T) {
			h := axisHarness(t)
			ctx := context.Background()
			in := axisInput()
			switch kind {
			case "add":
				in.AxisIDs = []string{"old", "new"}
				in.Mapping[0].Values = append(in.Mapping[0].Values, catalog.AttributeValue{AttributeID: "old", Value: "سیاه"})
				in.Mapping[1].Values = append(in.Mapping[1].Values, catalog.AttributeValue{AttributeID: "old", Value: "سفید"})
			case "remove-last":
				h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Variants = w.Variants[:2]; w.Offers = w.Offers[:3]; return nil })
				in.AxisIDs = []string{}
				in.Mapping = in.Mapping[:1]
				in.Mapping[0].Values = []catalog.AttributeValue{}
			case "empty":
				in.CategoryID = "gold"
				in.AxisIDs = []string{"new"}
				in.Mapping = nil
			case "over-limit":
				h.repo.Update(ctx, func(w *backoffice.Workspace) error {
					for i := 0; i < backoffice.MaxAxisProducts; i++ {
						w.Products = append(w.Products, catalog.Product{ID: fmt.Sprint("limit-", i), CategoryID: "regular"})
					}
					return nil
				})
			}
			v, e := h.s.PreviewAxisConversion(ctx, h.admin, in)
			if kind == "over-limit" {
				if !errors.Is(e, backoffice.ErrInvalid) {
					t.Fatal("unbounded transaction")
				}
				return
			}
			if e != nil || !v.Ready {
				t.Fatal(v, e)
			}
			p, e := h.s.PrepareAxisConversion(ctx, h.admin, in, v.Fingerprint, h.key())
			if e != nil {
				t.Fatal(e)
			}
			if _, e = h.s.ApplyAxisConversion(ctx, h.admin, p.ID, p.Fingerprint, true); e != nil {
				t.Fatal(e)
			}
			w, _ := h.repo.Snapshot(ctx)
			if kind == "add" && len(w.Variants[0].Values) != 2 {
				t.Fatal("added axis missing")
			}
			if kind == "remove-last" && len(w.Variants[0].Values) != 0 {
				t.Fatal("last axis remains")
			}
		})
	}
}
