package bootstrap_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"slices"
)

func placeHTTPTestPurchase(t *testing.T, h http.Handler, buyer string, offers []string) (backoffice.PurchaseView, backoffice.QuoteView) {
	t.Helper()
	cart := readJSON[backoffice.CartView](t, commerceHTTP(t, h, buyer, "", "GET", "/storefront/cart", nil, "", 200))
	for _, id := range offers {
		cart = readJSON[backoffice.CartView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: id, Quantity: 1, Version: cart.Version}, "ops-cart-"+id, 200))
	}
	address := readJSON[commerce.Address](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/addresses", backoffice.AddressInput{Recipient: "گیرنده ساختگی", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Phone: "09000000111", Province: "تهران", City: "تهران", Street: "نشانی ساختگی برای آزمون", PostalCode: "1234567890", HouseNumber: "12"}, "ops-address", 200))
	shipping := map[string]string{}
	for _, line := range cart.Items {
		shipping[line.StoreID] = "post"
	}
	q := readJSON[backoffice.QuoteView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/quotes", backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: shipping, PaymentMethodID: "local"}, "ops-quote", 200))
	p := readJSON[backoffice.PurchaseView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/checkout/purchases", backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, "ops-place", 200))
	return p, q
}

func TestOrderOperationsHTTPPostgresScopedDeliveryAndRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := commerceSeed()
	seed.Commerce.Settings.Shipping[0].DispatchIdentifier = "delivery"
	seed.Offers = append(seed.Offers, catalog.Offer{ID: "offer-two", ProductID: "product", VariantID: "variant", StoreID: "s2", Price: 1200000, Stock: 10, Version: 1})
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "test-order-seller", Name: "عملیات سفارش ساختگی", Version: 1, Scope: identity.ScopeStore, Permissions: []string{"orders.read", backoffice.OrdersOperate}})
	seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: "ops", StoreID: "s1", RoleID: "test-order-seller"})
	seed.StoreMemberships = slices.DeleteFunc(seed.StoreMemberships, func(m identity.StoreMembership) bool { return m.UserID == "ops" && m.StoreID == "s2" })
	seed.Access.Grants = slices.DeleteFunc(seed.Access.Grants, func(g identity.RoleGrant) bool { return g.UserID == "ops" && g.StoreID == "s2" })
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
	admin := login(t, h, "09000000001")
	seller := loginAt(t, h, "seller", "09000000002")["token"].(string)
	p, q := placeHTTPTestPurchase(t, h, buyer, []string{"offer", "offer-two"})
	id := p.OrderIDs[0]
	path := "/seller/orders/" + id + "/operations"
	get := func(token, path string) backoffice.OrderOperationsView {
		return readJSON[backoffice.OrderOperationsView](t, commerceHTTP(t, h, token, "", "GET", path, nil, "", 200))
	}
	v := get(seller, path)
	if v.Address != nil || len(v.AllowedActions) != 0 {
		t.Fatal("prepayment privacy/actions", v)
	}
	commerceHTTP(t, h, seller, "", "GET", "/seller/orders/"+p.OrderIDs[1]+"/operations", nil, "", 404)
	commerceHTTP(t, h, seller, "", "POST", path, backoffice.OrderOperationInput{Version: v.Version, Action: "start"}, "unpaid-start", 409)
	d := readJSON[backoffice.DepositView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "local", PurchaseID: p.ID, AmountToman: q.TotalToman, Date: time.Now().UTC().Format("2006-01-02"), Reference: "OPS-HTTP-FUND"}, "ops-funding", 200))
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/deposits/review", backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedReference: d.Reference, VerifiedAmountToman: d.AmountToman, VerifiedDate: d.Date}, "ops-funding-review", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "ops-payment", 200)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.Shipping[0].DispatchIdentifier = "tracking"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	for _, action := range []string{"start", "supply", "pack"} {
		v = get(seller, path)
		v = readJSON[backoffice.OrderOperationsView](t, commerceHTTP(t, h, seller, "", "POST", path, backoffice.OrderOperationInput{Version: v.Version, Action: action}, "ops-"+action, 200))
	}
	if v.Address == nil || v.Address.RecipientNationalID != "" || v.DispatchIdentifier != "delivery" {
		t.Fatal("address/policy snapshot", v)
	}
	commerceHTTP(t, h, seller, "", "POST", path, backoffice.OrderOperationInput{Version: v.Version, Action: "ship", AllItems: true}, "ops-no-code", 400)
	commerceHTTP(t, h, seller, "", "POST", path, backoffice.OrderOperationInput{OrderID: p.OrderIDs[1], Version: v.Version, Action: "ship", AllItems: true, Identifiers: []string{"CODE"}}, "ops-path-tamper", 400)
	v = readJSON[backoffice.OrderOperationsView](t, commerceHTTP(t, h, seller, "", "POST", path, backoffice.OrderOperationInput{Version: v.Version, Action: "ship", AllItems: true, Identifiers: []string{"۶۵۴۳۲۱"}}, "ops-ship", 200))
	if v.Status != "shipped" || v.Identifiers[0] != "654321" {
		t.Fatal(v)
	}
	receive := fmt.Sprintf("/storefront/account/orders/%s/shipments/%s/receive", p.ID, id)
	commerceHTTP(t, h, buyer, "", "POST", receive, backoffice.OrderOperationInput{Version: v.Version, Action: "correct", AllItems: true}, "ops-spoof", 400)
	commerceHTTP(t, h, buyer, "", "POST", fmt.Sprintf("/storefront/account/orders/wrong/shipments/%s/receive", id), backoffice.OrderOperationInput{Version: v.Version, AllItems: true}, "ops-wrong-parent", 404)
	v = readJSON[backoffice.OrderOperationsView](t, commerceHTTP(t, h, buyer, "", "POST", receive, backoffice.OrderOperationInput{Version: v.Version, AllItems: true}, "ops-receive", 200))
	if v.Status != "delivered" || v.OrderStatus != "completed" {
		t.Fatal(v)
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.EmptyWorkspace())
	if e != nil {
		t.Fatal(e)
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	buyer = loginAt(t, h, "storefront", "09000000005")["token"].(string)
	v = get(buyer, fmt.Sprintf("/storefront/account/orders/%s/shipments/%s", p.ID, id))
	if v.Status != "delivered" || len(v.Events) != 5 || v.Events[len(v.Events)-1].ActorKind != "customer" || v.DispatchIdentifier != "delivery" {
		t.Fatal("lost transition history/snapshot after restart", v)
	}
	for _, event := range v.Events {
		if event.Reason != "" {
			t.Fatal("internal operational reason leaked to customer")
		}
	}
	w, _ := reopened.Snapshot(ctx)
	if w.Commerce.Purchases[0].OrderState != "processing" || w.Orders[1].Status != "awaiting-supply" {
		t.Fatal("one shipment completed all")
	}
	for _, r := range w.Commerce.ShippingReservations {
		if r.Status != "committed" {
			t.Fatal("delivery released consumed capacity", r)
		}
	}
}
