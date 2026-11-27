package backoffice_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"strings"
	"testing"
)

func TestOrderWorkspaceServerFiltersBoundedRowsAliasesAndPrivacy(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	q := backoffice.OrderWorkspaceQuery{Size: 1, Store: "store1", Payment: "paid", Order: "placed", Fulfillment: "awaiting-supply", Search: "۱۰۰۰"}
	page, e := h.s.QueryOrderWorkspace(ctx, h.ops, q)
	if e != nil || page.Total != 1 || len(page.Items) != 1 || page.Items[0].Groups != 2 {
		t.Fatal(page, e)
	}
	raw, _ := json.Marshal(page)
	for _, secret := range []string{"credit_sources", "address", "receipt", "4242424242424242"} {
		if strings.Contains(string(raw), secret) {
			t.Fatal("list private data", secret)
		}
	}
	q.Order = "completed"
	page, e = h.s.QueryOrderWorkspace(ctx, h.ops, q)
	if e != nil || page.Total != 0 {
		t.Fatal("independent stage filter", page, e)
	}
	for _, alias := range []string{p.ID, p.Number, p.OrderIDs[0]} {
		detail, e := h.s.ReadOrderWorkspace(ctx, h.ops, alias)
		if e != nil || detail.Purchase == nil || detail.Purchase.ID != p.ID || detail.Permissions[backoffice.CommerceOrdersConfirm] || len(detail.Purchase.Deposits) > 0 || detail.Purchase.WalletBalanceToman != "" || len(detail.Purchase.CreditSources) > 0 || detail.Purchase.Quote.Destination.CardNumber != "" || len(detail.Purchase.Events) > 0 {
			t.Fatal("alias or privacy", detail, e)
		}
	}
	if _, e = h.s.QueryOrderWorkspace(ctx, h.buyer, backoffice.OrderWorkspaceQuery{Size: 20}); e != backoffice.ErrForbidden {
		t.Fatal(e)
	}
	if _, e = h.s.QueryOrderWorkspace(ctx, h.ops, backoffice.OrderWorkspaceQuery{Size: 101}); e != backoffice.ErrInvalid {
		t.Fatal(e)
	}
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Orders = append(w.Orders, orders.Order{ID: "legacy-preserved", StoreID: "store1", Status: "delivered", Payment: "paid", Items: []orders.Item{{ID: "old-item", Quantity: 1, UnitPrice: 1000}}})
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	detail, e := h.s.ReadOrderWorkspace(ctx, h.admin, "legacy-preserved")
	if e != nil || !detail.Legacy || detail.Purchase != nil {
		t.Fatal("legacy fabricated", detail, e)
	}
	page, e = h.s.QueryOrderWorkspace(ctx, h.admin, backoffice.OrderWorkspaceQuery{Size: 1, Page: 1})
	if e != nil || page.Total != 2 || len(page.Items) != 1 || !page.Items[0].Legacy {
		t.Fatal("legacy/pagination", page, e)
	}
}
