package backoffice_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"testing"
)

func TestShipmentQueueRequiresOwnProcessingAndScopesEveryProjection(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	one, two := p.OrderIDs[0], p.OrderIDs[1]
	query := backoffice.ShipmentQuery{Size: 20}
	empty, e := h.s.QueryShipments(ctx, h.admin, query)
	if e != nil || empty.Total != 0 {
		t.Fatal("paid must not imply processing", empty, e)
	}
	if _, e = h.s.ReadShipment(ctx, h.admin, one); e != backoffice.ErrNotFound {
		t.Fatal(e)
	}
	operationStep(t, h, h.admin, one, "start")
	page, e := h.s.QueryShipments(ctx, h.admin, query)
	if e != nil || page.Total != 1 || page.Items[0].ID != one {
		t.Fatal("parent start included other store", page, e)
	}
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: h.ops.ID, StoreID: "store1"}, identity.StoreMembership{UserID: h.ops.ID, StoreID: "store2"})
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "shipment-read", Scope: identity.ScopeStore, Permissions: []string{"orders.read"}}, identity.RoleDefinition{ID: "shipment-operate", Scope: identity.ScopeStore, Permissions: []string{backoffice.OrdersOperate}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: h.ops.ID, StoreID: "store1", RoleID: "shipment-read"}, identity.RoleGrant{UserID: h.ops.ID, StoreID: "store2", RoleID: "shipment-operate"})
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	seller := h.ops
	seller.Surface = identity.SurfaceSeller
	operationStep(t, h, h.admin, two, "start")
	page, e = h.s.QueryShipments(ctx, seller, query)
	if e != nil || page.Total != 1 || page.Items[0].ID != one || len(page.Stores) != 1 || page.Stores[0].ID != "store1" {
		t.Fatal("scope leaked", page, e)
	}
	detail, e := h.s.ReadShipment(ctx, seller, one)
	if e != nil || detail.Operations.Address != nil || len(detail.Operations.AllowedActions) != 0 {
		t.Fatal("cross-store action union or early address leak", detail, e)
	}
	if _, e = h.s.ReadShipment(ctx, seller, two); e != backoffice.ErrForbidden {
		t.Fatal("foreign read permission", e)
	}
	if _, e = h.s.TransitionOrder(ctx, seller, backoffice.OrderOperationInput{OrderID: one, Version: detail.Operations.Version, Action: "supply"}, h.key()); e != backoffice.ErrForbidden {
		t.Fatal("read acquired action from other store", e)
	}
	query.Store = "store2"
	page, e = h.s.QueryShipments(ctx, seller, query)
	if e != nil || page.Total != 0 {
		t.Fatal("foreign count", page, e)
	}
	query.Store = ""
	raw, _ := json.Marshal(page)
	for _, secret := range []string{"credit_sources", "recipient", "phone", "wallet_balance"} {
		if strings.Contains(string(raw), secret) {
			t.Fatal("private list", secret)
		}
	}
	if _, e = h.s.QueryShipments(ctx, h.buyer, query); e != backoffice.ErrForbidden {
		t.Fatal(e)
	}
	query.Size = 101
	if _, e = h.s.QueryShipments(ctx, h.admin, query); e != backoffice.ErrInvalid {
		t.Fatal(e)
	}
}

func TestShipmentTerminalHistorySurvivesCompletionAndCancellation(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	for _, id := range p.OrderIDs {
		operationStep(t, h, h.admin, id, "start")
	}
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Orders {
			if w.Orders[i].ID == p.OrderIDs[0] {
				w.Orders[i].Status = "delivered"
				w.Orders[i].OrderStatus = "completed"
			} else {
				w.Orders[i].Status = "cancelled"
				w.Orders[i].OrderStatus = "cancelled"
				w.Orders[i].Payment = "returned-to-wallet"
			}
		}
		w.Commerce.Purchases[0].OrderState = "completed"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	page, e := h.s.QueryShipments(ctx, h.admin, backoffice.ShipmentQuery{Size: 20})
	if e != nil || page.Total != 0 {
		t.Fatal("terminal in active queue", page, e)
	}
	page, e = h.s.QueryShipments(ctx, h.admin, backoffice.ShipmentQuery{Size: 1, Status: "all"})
	if e != nil || page.Total != 2 || len(page.Items) != 1 {
		t.Fatal("history disappeared", page, e)
	}
	page, e = h.s.QueryShipments(ctx, h.admin, backoffice.ShipmentQuery{Size: 20, Status: "cancelled"})
	if e != nil || page.Total != 1 || page.Items[0].ID != p.OrderIDs[1] {
		t.Fatal(page, e)
	}
	detail, e := h.s.ReadShipment(ctx, h.admin, p.OrderIDs[1])
	if e != nil || len(detail.Operations.AllowedActions) != 0 {
		t.Fatal("cancelled action enabled", detail, e)
	}
}
