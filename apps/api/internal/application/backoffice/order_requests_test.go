package backoffice_test

import (
	"bytes"
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"sync"
	"testing"
)

func createCase(t *testing.T, h *commerceHarness, p backoffice.PurchaseView, kind string, order int, qty int) backoffice.OrderRequestView {
	t.Helper()
	d, e := h.s.ReadOrderRequests(context.Background(), h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	var line backoffice.RequestItemView
	for _, v := range d.Items {
		if v.OrderID == p.OrderIDs[order] {
			line = v
			break
		}
	}
	b, e := h.s.CreateOrderRequest(context.Background(), h.buyer, backoffice.OrderRequestInput{PurchaseID: p.ID, Version: d.Version, Kind: kind, Reason: "درخواست ساختگی مشتری", Items: []backoffice.RequestedItemInput{{OrderID: line.OrderID, ItemID: line.ItemID, Quantity: qty}}}, h.key())
	return decodeCommerce[backoffice.OrderRequestView](t, b, e)
}
func reviewCase(t *testing.T, h *commerceHarness, r backoffice.OrderRequestView, action string, health *bool) backoffice.OrderRequestView {
	t.Helper()
	b, e := h.s.ReviewOrderRequest(context.Background(), h.admin, backoffice.OrderRequestReview{ID: r.ID, Version: r.Version, Action: action, Reason: "نتیجه ساختگی رسیدگی", Healthy: health}, h.key())
	return decodeCommerce[backoffice.OrderRequestView](t, b, e)
}

func TestOrderRequestsScopesPartialUnitsAndNoFinancialEffects(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	if _, e := h.s.ReadOrderRequests(ctx, h.other, p.ID); e != backoffice.ErrNotFound {
		t.Fatal(e)
	}
	d, e := h.s.ReadOrderRequests(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	line := d.Items[0]
	input := backoffice.OrderRequestInput{PurchaseID: p.ID, Version: d.Version, Kind: "cancel", Reason: "دلیل ساختگی", Items: []backoffice.RequestedItemInput{{OrderID: line.OrderID, ItemID: line.ItemID, Quantity: 2}}}
	if _, e = h.s.CreateOrderRequest(ctx, h.buyer, input, h.key()); e != backoffice.ErrConflict {
		t.Fatal("excess", e)
	}
	input.Items[0].Quantity = 1
	before, _ := h.repo.Snapshot(ctx)
	b, e := h.s.CreateOrderRequest(ctx, h.buyer, input, "create-once")
	r := decodeCommerce[backoffice.OrderRequestView](t, b, e)
	again, e := h.s.CreateOrderRequest(ctx, h.buyer, input, "create-once")
	if e != nil || !bytes.Equal(b, again) {
		t.Fatal("replay", e)
	}
	if _, e = h.s.CreateOrderRequest(ctx, h.buyer, input, "stale"); e != backoffice.ErrConflict {
		t.Fatal("version", e)
	}
	if _, e = h.s.ReviewOrderRequest(ctx, h.ops, backoffice.OrderRequestReview{ID: r.ID, Version: r.Version, Action: "approve", Reason: "x"}, h.key()); e != backoffice.ErrForbidden {
		t.Fatal("ops review", e)
	}
	r = reviewCase(t, h, r, "reject", nil)
	r = createCase(t, h, p, "cancel", 0, 1)
	r = reviewCase(t, h, r, "approve", nil)
	after, _ := h.repo.Snapshot(ctx)
	if !r.StockApplied || len(after.Commerce.Entries) != len(before.Commerce.Entries) || after.Offers[0].Stock != 5 || after.Offers[1].Stock != 4 {
		t.Fatal("wrong stock/money", r, after.Offers)
	}
	if after.Commerce.ShippingReservations[0].Status != "released" || after.Commerce.ShippingReservations[1].Status != "committed" {
		t.Fatal("capacity scope")
	}
	d, _ = h.s.ReadOrderRequests(ctx, h.buyer, p.ID)
	if d.Items[0].AvailableCancel != 0 {
		t.Fatal("units reused", d)
	}
	// Rejecting a case frees only its own original unit indices.
	if len(after.Commerce.OrderRequests) != 2 || after.Commerce.OrderRequests[0].Items[0].UnitIndices[0] != after.Commerce.OrderRequests[1].Items[0].UnitIndices[0] {
		t.Fatal("unit allocation")
	}
}

func TestReturnNeedsApprovalReceiptAndHealthyInspectionKeepsCompletedHistory(t *testing.T) {
	for _, receiveFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "approve-first", true: "receive-first"}[receiveFirst], func(t *testing.T) {
			h, p := paidOperationsHarness(t)
			ctx := context.Background()
			id := p.OrderIDs[0]
			for _, action := range []string{"start", "supply", "pack", "ship", "deliver"} {
				operationStep(t, h, h.admin, id, action)
			}
			r := createCase(t, h, p, "return", 0, 1)
			before, _ := h.repo.Snapshot(ctx)
			if receiveFirst {
				r = reviewCase(t, h, r, "receive", nil)
				r = reviewCase(t, h, r, "approve", nil)
			} else {
				r = reviewCase(t, h, r, "approve", nil)
				r = reviewCase(t, h, r, "receive", nil)
			}
			bad := false
			r = reviewCase(t, h, r, "inspect", &bad)
			after, _ := h.repo.Snapshot(ctx)
			if r.StockApplied || after.Offers[0].Stock != before.Offers[0].Stock {
				t.Fatal("bad returned to stock")
			}
			good := true
			r = reviewCase(t, h, r, "inspect", &good)
			after, _ = h.repo.Snapshot(ctx)
			if !r.StockApplied || after.Offers[0].Stock != before.Offers[0].Stock+1 || len(after.Commerce.Entries) != len(before.Commerce.Entries) {
				t.Fatal("healthy stock or money")
			}
			v, e := h.s.ReadOrderOperations(ctx, h.admin, id, "")
			if e != nil || v.Status != "delivered" || v.OrderStatus != "completed" {
				t.Fatal("history erased", v, e)
			}
		})
	}
}

func TestCancellationApprovalVersusShipmentAtomic(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	id := p.OrderIDs[0]
	for _, action := range []string{"start", "supply", "pack"} {
		operationStep(t, h, h.admin, id, action)
	}
	r := createCase(t, h, p, "cancel", 0, 1)
	v, _ := h.s.ReadOrderOperations(ctx, h.admin, id, "")
	var wg sync.WaitGroup
	results := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e := h.s.ReviewOrderRequest(ctx, h.admin, backoffice.OrderRequestReview{ID: r.ID, Version: r.Version, Action: "approve", Reason: "دلیل ساختگی"}, "race-approve")
		results <- e
	}()
	go func() {
		defer wg.Done()
		_, e := h.s.TransitionOrder(ctx, h.admin, backoffice.OrderOperationInput{OrderID: id, Version: v.Version, Action: "ship", AllItems: true, Identifiers: []string{"RACE"}}, "race-ship")
		results <- e
	}()
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else if e != backoffice.ErrConflict {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatal("both committed", success)
	}
}

func TestOrderRequestReplayRechecksRevokedPermissions(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	r := createCase(t, h, p, "cancel", 0, 1)
	input := backoffice.OrderRequestReview{ID: r.ID, Version: r.Version, Action: "reject", Reason: "ساختگی"}
	if _, e := h.s.ReviewOrderRequest(ctx, h.admin, input, "review-once"); e != nil {
		t.Fatal(e)
	}
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == h.admin.ID {
				w.Users[i].Active = false
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if _, e := h.s.ReviewOrderRequest(ctx, h.admin, input, "review-once"); e != backoffice.ErrForbidden {
		t.Fatal("revoked replay", e)
	}
}
