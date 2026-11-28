package backoffice_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

type activityMemory struct {
	mu     sync.Mutex
	events []operationlog.Event
}

func (m *activityMemory) Write(e operationlog.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}
func (m *activityMemory) Scan(ctx context.Context, q operationlog.Query, visit func(operationlog.Event)) error {
	m.mu.Lock()
	events := append([]operationlog.Event{}, m.events...)
	m.mu.Unlock()
	for _, e := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		if q.Match(e) {
			visit(e)
		}
	}
	return nil
}
func (m *activityMemory) count(name string) int {
	n := 0
	m.Scan(context.Background(), operationlog.Query{}, func(e operationlog.Event) {
		if e.Type == "behavior" && e.EventName == name {
			n++
		}
	})
	return n
}

func TestCustomerActivityUsesCommittedStateAndKeepsCheckoutOrigin(t *testing.T) {
	h := newCommerceHarness(t)
	buyer := h.other
	events := &activityMemory{}
	recorder := operationlog.New(events, "api", "test", "storefront")
	h.s.UseOperationLogs(recorder, events, nil)
	ctx := operationlog.WithRecorder(operationlog.WithContext(context.Background(), operationlog.ContextData{OperationID: "checkout-root", RequestID: "request-start", Device: "mobile"}), recorder)
	operationlog.SetActor(ctx, "storefront", buyer.ID, "")
	cart := h.cart(t, buyer)
	input := backoffice.CartMutation{Action: "add", OfferID: "o1", Quantity: 1, Version: cart.Version}
	// Fixture offer IDs are read from authoritative state, not guessed by client.
	state, _ := h.repo.Snapshot(ctx)
	input.OfferID = state.Offers[0].ID
	key := h.key()
	for i := 0; i < 2; i++ {
		if _, err := h.s.MutateCart(ctx, backoffice.CartOwner{Actor: buyer}, input, key); err != nil {
			t.Fatal(err)
		}
	}
	if events.count("cart.item-added") != 1 {
		t.Fatal("cart replay counted twice")
	}
	address := h.address(t, buyer)
	cart = h.cart(t, buyer)
	quoteInput := backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{state.Offers[0].StoreID: "post"}, PaymentMethodID: "card-transfer", WalletToman: "0"}
	raw, err := h.s.CreateQuote(ctx, buyer, quoteInput, h.key())
	quote := decodeCommerce[backoffice.QuoteView](t, raw, err)
	purchase := h.accept(t, buyer, quote)
	if events.count("checkout.started") != 1 || events.count("checkout.placed") != 1 || events.count("checkout.paid") != 0 || events.count("cart.item-removed") != 0 {
		t.Fatal("wrong placement/consumed cart events")
	}
	deposit := h.deposit(t, buyer, purchase.ID, "card-transfer", quote.TotalToman, "activity-fixture")
	h.approve(t, deposit)
	if events.count("checkout.paid") != 0 {
		t.Fatal("wallet top-up counted as order payment")
	}
	rows, err := h.s.ReadPurchases(ctx, buyer, purchase.ID)
	if err != nil || len(rows) != 1 {
		t.Fatal(err)
	}
	adminCtx := operationlog.WithRecorder(operationlog.WithContext(context.Background(), operationlog.ContextData{OperationID: "admin-confirm", RequestID: "request-confirm", Device: "desktop"}), recorder)
	operationlog.SetActor(adminCtx, "backoffice", h.admin.ID, "")
	confirmKey := h.key()
	confirm := backoffice.PurchaseAdminInput{ID: purchase.ID, Version: rows[0].Version}
	for i := 0; i < 2; i++ {
		if _, err := h.s.ConfirmPurchase(adminCtx, h.admin, confirm, confirmKey); err != nil {
			t.Fatal(err)
		}
	}
	if events.count("checkout.paid") != 1 {
		t.Fatal("paid replay counted twice")
	}
	var paid operationlog.Event
	events.Scan(ctx, operationlog.Query{Type: "behavior", EventName: "checkout.paid"}, func(e operationlog.Event) { paid = e })
	if paid.ActorID != h.admin.ID || paid.SubjectID != buyer.ID || paid.OperationID != "checkout-root" || paid.ParentOperationID != "admin-confirm" || paid.Attributes["device"] != "mobile" {
		t.Fatal("lost origin or confused actor/customer", paid)
	}
	// Simulate duplicate transport delivery. It must not inflate the report.
	events.Write(paid)
	funnel, err := h.s.ReadActivityFunnel(ctx, h.admin, operationlog.Query{Device: "mobile", CategoryID: state.Products[0].CategoryID})
	if err != nil || funnel.UniquePaidPurchases != 1 {
		t.Fatal(funnel, err)
	}
	if _, err = h.s.ReadActivityFunnel(ctx, buyer, operationlog.Query{}); err != backoffice.ErrForbidden {
		t.Fatal("customer analytics access", err)
	}
	current, _ := h.repo.Snapshot(ctx)
	public := backoffice.CatalogReadWorkspace(current)
	encoded, _ := json.Marshal(public)
	if strings.Contains(string(encoded), "activity_origins") {
		t.Fatal("private origins in public projection")
	}
}
