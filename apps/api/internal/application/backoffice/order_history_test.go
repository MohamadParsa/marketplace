package backoffice_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"testing"
)

func TestOrderHistoryPaginationPrivacyAndIndependentGroups(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: h.ops.ID, StoreID: "store1"})
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "history-seller", Scope: identity.ScopeStore, Permissions: []string{"orders.read", backoffice.OrdersOperate}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: h.ops.ID, StoreID: "store1", RoleID: "history-seller"})
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	for _, action := range []string{"start", "supply", "pack", "ship", "deliver"} {
		operationStep(t, h, h.admin, p.OrderIDs[0], action)
	}
	first, e := h.s.ReadOrderHistory(ctx, h.buyer, p.ID, 0, 2)
	if e != nil || len(first.Events) != 2 || first.NextBefore == 0 {
		t.Fatal(first, e)
	}
	all := map[string]bool{}
	before := 0
	count := 0
	for {
		page, e := h.s.ReadOrderHistory(ctx, h.buyer, p.ID, before, 2)
		if e != nil {
			t.Fatal(e)
		}
		for _, event := range page.Events {
			if all[event.ID] || event.ID == "" || event.Reason != "" || event.ActorName != "" {
				t.Fatal("history privacy/duplicates", event)
			}
			all[event.ID] = true
			count++
		}
		if page.NextBefore == 0 {
			break
		}
		before = page.NextBefore
	}
	state, _ := h.repo.Snapshot(ctx)
	if count != len(state.Commerce.Purchases[0].Events) {
		t.Fatal("missing history", count)
	}
	admin, e := h.s.ReadOrderHistory(ctx, h.admin, p.ID, 0, 50)
	if e != nil || admin.Events[0].Reason == "" {
		t.Fatal("private history unavailable", admin, e)
	}
	seller := h.ops
	seller.Surface = identity.SurfaceSeller
	scoped, e := h.s.ReadOrderHistory(ctx, seller, p.ID, 0, 50)
	if e != nil || len(scoped.Events) != 5 {
		t.Fatal(scoped, e)
	}
	for _, event := range scoped.Events {
		if event.OrderID != p.OrderIDs[0] || event.Reason != "" || event.ActorName != "" {
			t.Fatal("seller history", event)
		}
	}
	if _, e := h.s.ReadOrderHistory(ctx, h.other, p.ID, 0, 20); e != backoffice.ErrNotFound {
		t.Fatal("foreign history", e)
	}
	own, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil || own[0].OrderStatus != "processing" || own[0].PaymentStatus != "paid" || own[0].FulfillmentStatus != "mixed" {
		t.Fatal("parent state", own, e)
	}
	r := reviewCase(t, h, createCase(t, h, p, "cancel", 1, 1), "approve", nil)
	f := caseRefundInput(t, h, p, r, "wallet")
	b, e := h.s.RefundOrderRequest(ctx, h.finance, f, h.key())
	decodeCommerce[backoffice.OrderRefundSummary](t, b, e)
	latest, e := h.s.ReadOrderHistory(ctx, h.buyer, p.ID, 0, 50)
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(latest)
	if strings.Contains(string(raw), "تأیید مالی ساختگی") || strings.Contains(string(raw), "credit_allocations") {
		t.Fatal("private financial history")
	}
	again, e := h.s.ReadOrderHistory(ctx, h.buyer, p.ID, first.NextBefore, 50)
	if e != nil {
		t.Fatal(e)
	}
	for _, event := range again.Events {
		if event.Sequence >= first.NextBefore {
			t.Fatal("cursor moved after append")
		}
	}
	// The cancellation's operational event must not leak its support reason in the seller panel either.
	r = createCase(t, h, p, "return", 0, 1)
	r = reviewCase(t, h, r, "approve", nil)
	healthy := true
	_ = reviewCase(t, h, r, "receive", &healthy)
	operation, e := h.s.ReadOrderOperations(ctx, seller, p.OrderIDs[0], "")
	if e != nil {
		t.Fatal(e)
	}
	for _, event := range operation.Events {
		if strings.Contains(event.Action, "stock-returned") && event.Reason != "" {
			t.Fatal("private case reason in operations")
		}
	}
}
