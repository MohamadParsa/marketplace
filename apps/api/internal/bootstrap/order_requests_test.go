package bootstrap_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"testing"
	"time"
)

func TestOrderRequestsHTTPPostgresPrivacyAndRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		*w = commerceSeed()
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "request-seller", Name: "نقش ساختگی مشاهده سفارش", Scope: identity.ScopeStore, Version: 1, Permissions: []string{"orders.read"}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "ops", StoreID: "s1", RoleID: "request-seller"})
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
	admin := login(t, h, "09000000001")
	seller := loginAt(t, h, "seller", "09000000002")["token"].(string)
	p, q := placeHTTPTestPurchase(t, h, buyer, []string{"offer"})
	d := readJSON[backoffice.DepositView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "local", PurchaseID: p.ID, AmountToman: q.TotalToman, Date: time.Now().UTC().Format("2006-01-02"), Reference: "REQUEST-FUND"}, "request-fund", 200))
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/deposits/review", backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedReference: d.Reference, VerifiedAmountToman: d.AmountToman, VerifiedDate: d.Date}, "request-fund-review", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "request-pay", 200)
	path := "/storefront/account/orders/" + p.ID + "/requests"
	dir := readJSON[backoffice.OrderRequestDirectory](t, commerceHTTP(t, h, buyer, "", "GET", path, nil, "", 200))
	item := dir.Items[0]
	input := backoffice.OrderRequestInput{Version: dir.Version, Kind: "cancel", Reason: "دلیل خصوصی ساختگی مشتری", Items: []backoffice.RequestedItemInput{{OrderID: item.OrderID, ItemID: item.ItemID, Quantity: 1}}}
	commerceHTTP(t, h, seller, "", "POST", "/seller/purchases/"+p.ID+"/requests", input, "seller-create", 403)
	r := readJSON[backoffice.OrderRequestView](t, commerceHTTP(t, h, buyer, "", "POST", path, input, "request-create", 200))
	commerceHTTP(t, h, buyer, "", "POST", path, input, "request-create", 200)
	commerceHTTP(t, h, buyer, "", "POST", path, input, "request-stale", 409)
	scoped := readJSON[backoffice.OrderRequestDirectory](t, commerceHTTP(t, h, seller, "", "GET", "/seller/purchases/"+p.ID+"/requests", nil, "", 200))
	if len(scoped.Requests) != 1 || scoped.Requests[0].Reason != "" || scoped.CanCreate || scoped.CanReview {
		t.Fatal("seller privacy", scoped)
	}
	review := backoffice.OrderRequestReview{Version: r.Version, Action: "approve", Reason: "نتیجه ساختگی رسیدگی"}
	commerceHTTP(t, h, buyer, "", "POST", "/backoffice/order-requests/"+r.ID+"/review", review, "customer-review", 401)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/order-requests/"+r.ID+"/review", review, "request-approve", 200)
	state, _ := repo.Snapshot(ctx)
	if len(state.Commerce.OrderRequests) != 1 || !state.Commerce.OrderRequests[0].StockApplied || len(state.Commerce.Entries) != 2 {
		t.Fatal("atomic case stock/money")
	}
	reloaded, err := reviewpostgres.New(ctx, db, commerceSeed())
	if err != nil {
		t.Fatal(err)
	}
	restarted := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reloaded)
	buyer = loginAt(t, restarted, "storefront", "09000000005")["token"].(string)
	persisted := readJSON[backoffice.OrderRequestDirectory](t, commerceHTTP(t, restarted, buyer, "", "GET", path, nil, "", 200))
	if len(persisted.Requests) != 1 || persisted.Requests[0].Status != "approved" {
		t.Fatal("restart", persisted)
	}
}

