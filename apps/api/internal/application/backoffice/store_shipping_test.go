package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"reflect"
	"sync"
	"testing"
)

func shippingSeller(h *commerceHarness) identity.User {
	a := h.buyer
	a.Surface = identity.SurfaceSeller
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: a.ID, StoreID: "store1"})
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "ship-test", Scope: identity.ScopeStore, Permissions: []string{backoffice.StoresShippingRead, backoffice.StoresShippingWrite}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: a.ID, StoreID: "store1", RoleID: "ship-test"})
		return nil
	})
	return a
}
func shippingCommand(action string, version int) backoffice.StoreShippingCommand {
	return backoffice.StoreShippingCommand{Action: action, MethodID: "post", Stores: []backoffice.StoreShippingTarget{{ID: "store1", Version: version}}}
}
func TestStoreShippingApprovalScopesVersionsRaceAndPublicSafety(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	seller := shippingSeller(h)
	rows, err := h.s.ReadStoreShipping(ctx, seller)
	if err != nil || len(rows.Stores) != 1 || rows.Stores[0].CanReview || !rows.Stores[0].CanWrite {
		t.Fatal(rows, err)
	}
	before, _ := h.repo.Snapshot(ctx)
	in := shippingCommand("draft", 0)
	method := methodInput("post")
	method.Name = "پست نمونه"
	method.FeeToman = "123"
	in.Method = &method
	for _, actor := range []identity.User{h.buyer, h.finance} {
		if _, err = h.s.SaveStoreShipping(ctx, actor, in, "denied"); !errors.Is(err, backoffice.ErrForbidden) {
			t.Fatal("permission", err)
		}
	}
	other := in
	other.Stores = []backoffice.StoreShippingTarget{{ID: "store2"}}
	if _, err = h.s.SaveStoreShipping(ctx, seller, other, "other"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("other store", err)
	}
	save := func(actor identity.User, c backoffice.StoreShippingCommand, key string) {
		t.Helper()
		if _, e := h.s.SaveStoreShipping(ctx, actor, c, key); e != nil {
			t.Fatal(c.Action, e)
		}
	}
	save(seller, in, "draft")
	save(seller, in, "draft")
	w, _ := h.repo.Snapshot(ctx)
	if _, ok := backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0]); ok {
		t.Fatal("draft published")
	}
	if !reflect.DeepEqual(before.Commerce.Entries, w.Commerce.Entries) || !reflect.DeepEqual(before.Orders, w.Orders) {
		t.Fatal("protected data")
	}
	submit := shippingCommand("submit", 1)
	save(seller, submit, "submit")
	approve := shippingCommand("approve", 2)
	approve.Reason = "بررسی تفاوت تعرفه و پوشش"
	if _, err = h.s.SaveStoreShipping(ctx, seller, approve, "seller-review"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("seller approved", err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i, action := range []string{"approve", "reject"} {
		wg.Add(1)
		go func(i int, action string) {
			defer wg.Done()
			command := approve
			command.Action = action
			_, e := h.s.SaveStoreShipping(ctx, h.admin, command, fmt.Sprintf("review-%d", i))
			results <- e
		}(i, action)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for e := range results {
		if e == nil {
			success++
		} else if errors.Is(e, backoffice.ErrConflict) {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("racing review", success, conflict)
	}
	// Create a deterministic next revision regardless of which reviewer won.
	in.Stores[0].Version = 3
	save(seller, in, "draft-2")
	save(seller, shippingCommand("submit", 4), "submit-2")
	approve.Stores[0].Version = 5
	save(h.admin, approve, "approve-2")
	save(h.admin, approve, "approve-2")
	w, _ = h.repo.Snapshot(ctx)
	effective, ok := backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0])
	if !ok || commerce.Toman(effective.Fee) != "123" || effective.StoreRevision != 2 {
		t.Fatal("live", effective, ok)
	}
	h.add(t, h.buyer, "offer1", 1)
	quote := h.quote(t, h.buyer, "wallet")
	if quote.ShippingToman != "123" {
		t.Fatal("checkout did not resolve custom shipping", quote.ShippingToman)
	}
	in.Stores[0].Version = 6
	method.FeeToman = "456"
	save(seller, in, "draft-3")
	w, _ = h.repo.Snapshot(ctx)
	effective, ok = backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0])
	if !ok || commerce.Toman(effective.Fee) != "123" {
		t.Fatal("pending overwrote live")
	}
	// The public method remains a current upper bound; historical snapshots are values.
	old := effective
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].Active = false
		w.Commerce.Settings.Shipping[0].Version++
		return nil
	})
	w, _ = h.repo.Snapshot(ctx)
	if _, ok = backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0]); ok {
		t.Fatal("inactive bypass")
	}
	if !old.Active || commerce.Toman(old.Fee) != "123" {
		t.Fatal("historical changed")
	}
	if _, err = h.s.SaveStoreShipping(ctx, seller, shippingCommand("submit", 7), "stale-base"); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("stale base", err)
	}
	// Revoking permissions is also enforced for a replay of an old key.
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Access.Grants = nil; return nil })
	if _, err = h.s.SaveStoreShipping(ctx, seller, in, "draft-3"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("revoked retry", err)
	}
}

func TestStoreShippingCoverageIntersectionBatchAndRebase(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		m := &w.Commerce.Settings.Shipping[0]
		m.AllCities = false
		m.CityIDs = []string{"tehran-city"}
		m.CategoryMode = "selected"
		m.AllowedCategories = []string{"regular"}
		return nil
	})
	method := methodInput("post")
	method.Name = "پست نمونه"
	in := shippingCommand("draft", 0)
	in.Method = &method
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, in, "wide-city"); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("city widening", err)
	}
	method.AllCities = false
	method.CityIDs = []string{"tehran-city"}
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, in, "wide-category"); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("category widening", err)
	}
	method.CategoryMode = "selected"
	method.AllowedCategories = []string{"regular"}
	method.CollectionMode = "collect"
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, in, "collection"); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("collection changed", err)
	}
	method.CollectionMode = "prepaid"
	in.All = true
	in.Stores = append(in.Stores, backoffice.StoreShippingTarget{ID: "store2", Version: 1})
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, in, "batch-stale"); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("stale batch", err)
	}
	w, _ := h.repo.Snapshot(ctx)
	if len(w.Commerce.StoreShipping) != 0 {
		t.Fatal("partial batch")
	}
	in.Stores[1].Version = 0
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, in, "batch"); err != nil {
		t.Fatal(err)
	}
	w, _ = h.repo.Snapshot(ctx)
	if len(w.Commerce.StoreShipping) != 2 {
		t.Fatal("all missing")
	}
	submit := shippingCommand("submit", 1)
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, submit, "submit"); err != nil {
		t.Fatal(err)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Commerce.Settings.Shipping[0].Version++; return nil })
	approve := shippingCommand("approve", 2)
	approve.BaseVersion = 1
	approve.Reason = "تأیید آزمون"
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, approve, "stale-review"); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("base changed", err)
	}
}

func TestStoreShippingReviewPermissionIsSeparateAndCollectionChangesNeedReview(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "shipping-review-only", Scope: identity.ScopeGlobal, Permissions: []string{backoffice.StoresRead, backoffice.StoresShippingRead, backoffice.StoresShippingReview}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: h.finance.ID, RoleID: "shipping-review-only"})
		return nil
	})
	rows, err := h.s.ReadStoreShipping(ctx, h.finance)
	if err != nil || !rows.Stores[0].CanReview || rows.Stores[0].CanWrite {
		t.Fatal("review rights", rows, err)
	}
	method := methodInput("post")
	method.Name = "پست نمونه"
	in := shippingCommand("draft", 0)
	in.Method = &method
	if _, err = h.s.SaveStoreShipping(ctx, h.finance, in, "reviewer-draft"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("reviewer gained edit", err)
	}
	if _, err = h.s.SaveStoreShipping(ctx, h.admin, in, "draft"); err != nil {
		t.Fatal(err)
	}
	if _, err = h.s.SaveStoreShipping(ctx, h.admin, shippingCommand("submit", 1), "submit"); err != nil {
		t.Fatal(err)
	}
	review := shippingCommand("approve", 2)
	review.Reason = "تفاوت نسخه بررسی شد"
	if _, err = h.s.SaveStoreShipping(ctx, h.finance, review, "review"); err != nil {
		t.Fatal(err)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].CollectionMode = "collect"
		w.Commerce.Settings.Shipping[0].Fee = 0
		w.Commerce.Settings.Shipping[0].Version++
		return nil
	})
	w, _ := h.repo.Snapshot(ctx)
	if _, ok := backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0]); ok {
		t.Fatal("collection change silently became free/prepaid")
	}
	in.BaseVersion = 1
	in.Stores[0].Version = 3
	method.CollectionMode = "collect"
	method.FeeToman = ""
	if _, err = h.s.SaveStoreShipping(ctx, h.admin, in, "rebase"); err != nil {
		t.Fatal(err)
	}
	submit := shippingCommand("submit", 4)
	submit.BaseVersion = 1
	if _, err = h.s.SaveStoreShipping(ctx, h.admin, submit, "submit-rebase"); err != nil {
		t.Fatal(err)
	}
	review.BaseVersion = 1
	review.Stores[0].Version = 5
	if _, err = h.s.SaveStoreShipping(ctx, h.finance, review, "review-rebase"); err != nil {
		t.Fatal(err)
	}
	w, _ = h.repo.Snapshot(ctx)
	effective, ok := backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0])
	if !ok || effective.CollectionMode != "collect" || effective.Fee != 0 {
		t.Fatal(effective, ok)
	}
}
