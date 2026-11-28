package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestOrderRefundHTTPPostgresConcurrentHoldAndRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = commerceSeed(); return nil }); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
	admin := login(t, h, "09000000001")
	p, q := placeHTTPTestPurchase(t, h, buyer, []string{"offer"})
	d := readJSON[backoffice.DepositView](t, commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "local", PurchaseID: p.ID, AmountToman: q.TotalToman, Date: time.Now().UTC().Format("2006-01-02"), Reference: "REFUND-FUND"}, "refund-fund", 200))
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/deposits/review", backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedReference: d.Reference, VerifiedAmountToman: d.AmountToman, VerifiedDate: d.Date}, "refund-fund-review", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "refund-pay", 200)
	path := "/storefront/account/orders/" + p.ID + "/requests"
	dir := readJSON[backoffice.OrderRequestDirectory](t, commerceHTTP(t, h, buyer, "", "GET", path, nil, "", 200))
	line := dir.Items[0]
	r := readJSON[backoffice.OrderRequestView](t, commerceHTTP(t, h, buyer, "", "POST", path, backoffice.OrderRequestInput{Version: dir.Version, Kind: "cancel", Reason: "ساختگی", Items: []backoffice.RequestedItemInput{{OrderID: line.OrderID, ItemID: line.ItemID, Quantity: 1}}}, "refund-case", 200))
	r = readJSON[backoffice.OrderRequestView](t, commerceHTTP(t, h, admin, "", "POST", "/backoffice/order-requests/"+r.ID+"/review", backoffice.OrderRequestReview{Version: r.Version, Action: "approve", Reason: "تأیید ساختگی"}, "refund-case-review", 200))
	input := backoffice.OrderRefundInput{Version: r.Version, Method: "bank", AcceptedAmountToman: q.TotalToman, Instructions: "مقصد خصوصی ساختگی", Reason: "تأیید مالی ساختگی"}
	target := "/backoffice/order-requests/" + r.ID + "/refund"
	commerceHTTP(t, h, buyer, "", "POST", target, input, "buyer-refund", 401)
	body, _ := json.Marshal(input)
	results := make(chan int, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for _, key := range []string{"refund-race-a", "refund-race-b"} {
		go func(k string) {
			defer wg.Done()
			req := httptest.NewRequest("POST", target, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+admin)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", k)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			results <- rec.Code
		}(key)
	}
	wg.Wait()
	close(results)
	ok, conflict := 0, 0
	for status := range results {
		switch status {
		case 200:
			ok++
		case 409:
			conflict++
		default:
			t.Fatal("unexpected race status", status)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatal(ok, conflict)
	}
	state, _ := repo.Snapshot(ctx)
	if len(state.Commerce.OrderRefunds) != 1 || len(state.Commerce.Refunds) != 1 || len(state.Commerce.Entries) != 3 {
		t.Fatal("non atomic or duplicate financial rows")
	}
	reloaded, e := reviewpostgres.New(ctx, db, commerceSeed())
	if e != nil {
		t.Fatal(e)
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reloaded)
	buyer = loginAt(t, h, "storefront", "09000000005")["token"].(string)
	admin = login(t, h, "09000000001")
	wallet := readJSON[backoffice.WalletView](t, commerceHTTP(t, h, buyer, "", "GET", "/storefront/wallet", nil, "", 200))
	if wallet.HeldToman != q.TotalToman || wallet.AvailableToman != "0" {
		t.Fatal("bank amount spendable after restart", wallet)
	}
	dir = readJSON[backoffice.OrderRequestDirectory](t, commerceHTTP(t, h, buyer, "", "GET", path, nil, "", 200))
	if dir.Requests[0].Refund == nil || dir.Requests[0].Refund.Status != "bank-pending" {
		t.Fatal(dir)
	}
	withdrawal := state.Commerce.Refunds[0]
	resolve := backoffice.RefundResolveInput{ID: withdrawal.ID, Version: withdrawal.Version, Complete: true, Reference: "BANK-PG-138", Date: time.Now().UTC().Format("2006-01-02"), Reason: "ثبت ساختگی"}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/refunds/resolve", resolve, "bank-no-amount", 400)
	resolve.VerifiedAmountToman = q.TotalToman
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/refunds/resolve", resolve, "bank-pg-complete", 200)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/refunds/resolve", resolve, "bank-pg-complete", 200)
	wallet = readJSON[backoffice.WalletView](t, commerceHTTP(t, h, buyer, "", "GET", "/storefront/wallet", nil, "", 200))
	if wallet.BalanceToman != "0" || wallet.HeldToman != "0" {
		t.Fatal("bank duplicated", wallet)
	}
}
