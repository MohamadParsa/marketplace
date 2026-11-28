package backoffice_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func paidOperationsHarness(t *testing.T) (*commerceHarness, backoffice.PurchaseView) {
	t.Helper()
	h := newCommerceHarness(t)
	ctx := context.Background()
	if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].DispatchIdentifier = "tracking"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "order-operations-funding")
	h.approve(t, d)
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	return h, p
}

func operationStep(t *testing.T, h *commerceHarness, a identity.User, id, action string) backoffice.OrderOperationsView {
	t.Helper()
	v, e := h.s.ReadOrderOperations(context.Background(), a, id, "")
	if e != nil {
		t.Fatal(e)
	}
	b, e := h.s.TransitionOrder(context.Background(), a, backoffice.OrderOperationInput{OrderID: id, Version: v.Version, Action: action, Reason: "ثبت مرحله ساختگی برای آزمون", Identifiers: []string{"POST-TEST-001"}, AllItems: true}, h.key())
	return decodeCommerce[backoffice.OrderOperationsView](t, b, e)
}

func TestOrderOperationsScopesPrivacyDispatchAndCompletion(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	one, two := p.OrderIDs[0], p.OrderIDs[1]
	if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: h.ops.ID, StoreID: "store1"})
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "test-order-operator", Scope: identity.ScopeStore, Permissions: []string{"orders.read", backoffice.OrdersOperate}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: h.ops.ID, StoreID: "store1", RoleID: "test-order-operator"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	seller := h.ops
	seller.Surface = identity.SurfaceSeller
	v, e := h.s.ReadOrderOperations(ctx, seller, one, "")
	if e != nil {
		t.Fatal(e)
	}
	if v.Address != nil || v.OrderStatus != "placed" || v.PaymentStatus != "paid" {
		t.Fatal("early customer disclosure or invented processing", v)
	}
	if _, e = h.s.ReadOrderOperations(ctx, seller, two, ""); e != backoffice.ErrNotFound {
		t.Fatal("foreign store read", e)
	}
	if _, e = h.s.TransitionOrder(ctx, h.finance, backoffice.OrderOperationInput{OrderID: one, Version: v.Version, Action: "start"}, h.key()); e != backoffice.ErrForbidden {
		t.Fatal("finance gained operations", e)
	}
	start := backoffice.OrderOperationInput{OrderID: one, Version: v.Version, Action: "start"}
	first, e := h.s.TransitionOrder(ctx, seller, start, "start-once")
	if e != nil {
		t.Fatal(e)
	}
	again, e := h.s.TransitionOrder(ctx, seller, start, "start-once")
	if e != nil || !bytes.Equal(first, again) {
		t.Fatal("operation replay", e)
	}
	v = decodeCommerce[backoffice.OrderOperationsView](t, first, nil)
	if v.OrderStatus != "processing" || v.Address != nil || len(v.Events) != 1 || v.Events[0].ActorKind != "seller" {
		t.Fatal("start state/privacy/history", v)
	}
	if v.Events[0].CustomerDescription == "" {
		t.Fatal("missing customer-safe step description")
	}
	if _, e = h.s.TransitionOrder(ctx, seller, start, "start-stale-version"); e != backoffice.ErrConflict {
		t.Fatal("stale order version", e)
	}
	v = operationStep(t, h, seller, one, "supply")
	if v.Address == nil || v.Address.Phone == "" || v.Address.RecipientNationalID != "" || v.Address.UserID != "" {
		t.Fatal("shipping privacy after supply", v.Address)
	}
	v = operationStep(t, h, seller, one, "pack")
	bad := backoffice.OrderOperationInput{OrderID: one, Version: v.Version, Action: "ship", AllItems: true}
	if _, e = h.s.TransitionOrder(ctx, seller, bad, h.key()); e != backoffice.ErrInvalid {
		t.Fatal("missing tracking accepted", e)
	}
	bad.Identifiers = []string{"۱۲۳", "123"}
	if _, e = h.s.TransitionOrder(ctx, seller, bad, h.key()); e != backoffice.ErrInvalid {
		t.Fatal("duplicate normalized tracking accepted", e)
	}
	bad.Identifiers = []string{"A\u0001B"}
	if _, e = h.s.TransitionOrder(ctx, seller, bad, h.key()); e != backoffice.ErrInvalid {
		t.Fatal("control character tracking accepted", e)
	}
	bad.Identifiers = []string{"۱۲۳۴", "TEST-2"}
	bad.AllItems = false
	if _, e = h.s.TransitionOrder(ctx, seller, bad, h.key()); e != backoffice.ErrInvalid {
		t.Fatal("partial dispatch accepted", e)
	}
	bad.AllItems = true
	b, e := h.s.TransitionOrder(ctx, seller, bad, h.key())
	v = decodeCommerce[backoffice.OrderOperationsView](t, b, e)
	if v.Status != "shipped" || len(v.Identifiers) != 2 || v.Identifiers[0] != "1234" {
		t.Fatal("dispatch snapshot", v)
	}
	if _, e = h.s.TransitionOrder(ctx, seller, backoffice.OrderOperationInput{OrderID: one, Version: v.Version, Action: "deliver", AllItems: true}, h.key()); e != backoffice.ErrForbidden {
		t.Fatal("seller self-confirmed customer delivery", e)
	}
	if _, e = h.s.ReadOrderOperations(ctx, h.other, one, p.ID); e != backoffice.ErrNotFound {
		t.Fatal("foreign customer shipment read", e)
	}
	v = operationStep(t, h, h.buyer, one, "deliver")
	if v.OrderStatus != "completed" || v.Status != "delivered" {
		t.Fatal(v)
	}
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil || rows[0].OrderStatus != "processing" {
		t.Fatal("one group completed parent", rows, e)
	}
	before, _ := h.repo.Snapshot(ctx)
	b, e = h.s.TransitionOrder(ctx, h.admin, backoffice.OrderOperationInput{OrderID: one, Version: v.Version, Action: "correct", Reason: "اصلاح تایپی کد با حفظ تاریخچه", Identifiers: []string{"CORRECTED"}}, h.key())
	decodeCommerce[backoffice.OrderOperationsView](t, b, e)
	after, _ := h.repo.Snapshot(ctx)
	event := after.Commerce.Purchases[0].Events[len(after.Commerce.Purchases[0].Events)-1]
	if len(event.PreviousIdentifiers) != 2 || event.PreviousIdentifiers[0] != "1234" || event.Identifiers[0] != "CORRECTED" || after.Orders[0].Status != "delivered" || len(before.Commerce.Entries) != len(after.Commerce.Entries) {
		t.Fatal("correction rewrote history/finance", event)
	}
	for _, action := range []string{"start", "supply", "pack", "ship", "deliver"} {
		operationStep(t, h, h.admin, two, action)
	}
	rows, e = h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil || rows[0].OrderStatus != "completed" {
		t.Fatal("all delivered did not complete parent", rows, e)
	}
}

func TestOrderOperationReplayRechecksRevokedPermission(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "only-order", Scope: identity.ScopeGlobal, Permissions: []string{backoffice.CommerceOrdersRead, backoffice.OrdersOperate}})
		w.Users[1].Role = ""
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: w.Users[1].ID, RoleID: "only-order"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	a := h.ops
	v, e := h.s.ReadOrderOperations(ctx, a, p.OrderIDs[0], "")
	if e != nil {
		t.Fatal(e)
	}
	in := backoffice.OrderOperationInput{OrderID: v.ID, Version: v.Version, Action: "start"}
	if _, e = h.s.TransitionOrder(ctx, a, in, "revoked-replay"); e != nil {
		t.Fatal(e)
	}
	if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Access.Roles[len(w.Access.Roles)-1].Permissions = []string{backoffice.CommerceOrdersRead}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, e = h.s.TransitionOrder(ctx, a, in, "revoked-replay"); e != backoffice.ErrForbidden {
		t.Fatal("idempotency bypassed permission revocation", e)
	}
}

func TestConcurrentDispatchAndCancellationCannotBothCommit(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	id := p.OrderIDs[0]
	for _, action := range []string{"start", "supply", "pack"} {
		operationStep(t, h, h.admin, id, action)
	}
	v, _ := h.s.ReadOrderOperations(ctx, h.admin, id, "")
	rows, _ := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	results := make(chan error, 2)
	go func() {
		_, e := h.s.TransitionOrder(ctx, h.admin, backoffice.OrderOperationInput{OrderID: id, Version: v.Version, Action: "ship", AllItems: true, Identifiers: []string{"RACE-POST"}}, "race-ship")
		results <- e
	}()
	go func() {
		_, e := h.s.AdminCancelPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: rows[0].Version, Reason: "لغو هم‌زمان ساختگی"}, "race-cancel")
		results <- e
	}()
	one, two := <-results, <-results
	if !((one == nil && two == backoffice.ErrConflict) || (two == nil && one == backoffice.ErrConflict)) {
		t.Fatal("dispatch/cancel race", one, two)
	}
	w, _ := h.repo.Snapshot(ctx)
	if w.Commerce.Purchases[0].Status == "cancelled" && w.Orders[0].Status == "shipped" {
		t.Fatal("cancelled purchase with committed shipment")
	}
}

func TestOrderSupplyHoldAndExplicitLegacyDispatchPolicy(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	id := p.OrderIDs[0]
	operationStep(t, h, h.admin, id, "start")
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Orders[0].Status = "on-hold"; return nil }); e != nil {
		t.Fatal(e)
	}
	v, _ := h.s.ReadOrderOperations(ctx, h.admin, id, "")
	if _, e := h.s.TransitionOrder(ctx, h.admin, backoffice.OrderOperationInput{OrderID: id, Version: v.Version, Action: "supply"}, h.key()); e != backoffice.ErrConflict {
		t.Fatal("supply hold bypass", e)
	}
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Orders[0].Status = "awaiting-supply"; return nil }); e != nil {
		t.Fatal(e)
	}
	operationStep(t, h, h.admin, id, "supply")
	operationStep(t, h, h.admin, id, "pack")
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Orders[0].ShippingTerms.DispatchIdentifier = ""; return nil }); e != nil {
		t.Fatal(e)
	}
	v, _ = h.s.ReadOrderOperations(ctx, h.admin, id, "")
	if _, e := h.s.TransitionOrder(ctx, h.admin, backoffice.OrderOperationInput{OrderID: id, Version: v.Version, Action: "ship", AllItems: true, Identifiers: []string{"LEGACY"}}, h.key()); e != backoffice.ErrConflict {
		t.Fatal("guessed dispatch policy", e)
	}
	in := backoffice.OrderOperationInput{OrderID: id, Version: v.Version, Action: "configure-dispatch", DispatchIdentifier: "delivery"}
	if _, e := h.s.TransitionOrder(ctx, h.admin, in, h.key()); e != backoffice.ErrInvalid {
		t.Fatal("policy configured without reason", e)
	}
	in.Reason = "تعیین صریح روش قدیمی"
	b, e := h.s.TransitionOrder(ctx, h.admin, in, h.key())
	v = decodeCommerce[backoffice.OrderOperationsView](t, b, e)
	if v.DispatchIdentifier != "delivery" {
		t.Fatal(v)
	}
	operationStep(t, h, h.admin, id, "ship")
}

func TestStoreCannotOverrideInheritedDispatchIdentifier(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	seller := shippingSeller(h)
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].DispatchIdentifier = "tracking"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	settings, e := h.s.ReadCommerceSettings(ctx, h.admin)
	if e != nil {
		t.Fatal(e)
	}
	method := settings.Shipping[0]
	method.DispatchIdentifier = "delivery"
	in := shippingCommand("draft", 0)
	in.Method = &method
	if _, e = h.s.SaveStoreShipping(ctx, seller, in, "invalid-dispatch-override"); e == nil {
		t.Fatal("store changed inherited dispatch identifier")
	}
	method.DispatchIdentifier = "tracking"
	if _, e = h.s.SaveStoreShipping(ctx, seller, in, "dispatch-draft"); e != nil {
		t.Fatal(e)
	}
	if _, e = h.s.SaveStoreShipping(ctx, seller, shippingCommand("submit", 1), "dispatch-submit"); e != nil {
		t.Fatal(e)
	}
	approval := shippingCommand("approve", 2)
	approval.Reason = "تأیید روش ساختگی"
	if _, e = h.s.SaveStoreShipping(ctx, h.admin, approval, "dispatch-approve"); e != nil {
		t.Fatal(e)
	}
	w, _ := h.repo.Snapshot(ctx)
	effective, ok := backoffice.EffectiveStoreShipping(w, "store1", w.Commerce.Settings.Shipping[0])
	if !ok || effective.DispatchIdentifier != "tracking" {
		t.Fatal("dispatch inheritance", effective, ok)
	}
}
