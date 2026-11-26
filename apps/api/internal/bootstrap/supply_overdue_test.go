package bootstrap_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

func TestSupplyOverduePostgresHTTPBoundaryScopesRestartAndRepeatedChecks(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := httpStoreShippingSeed(t)
	for i := range seed.Access.Roles {
		if seed.Access.Roles[i].ID == "store-shipping-test" {
			seed.Access.Roles[i].Permissions = append(seed.Access.Roles[i].Permissions, "orders.read")
		}
	}
	source := time.Date(2026, 11, 25, 11, 0, 0, 0, shipping.Tehran())
	snapshot, err := shipping.ResolvePreparation(source, 1, shipping.PreparationPolicy{DayBasis: shipping.WorkingDays, CalendarMode: shipping.Inherit}, shipping.Calendar{Version: 3, WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{"2026-11-28"}}, shipping.DefaultTiming())
	if err != nil {
		t.Fatal(err)
	}
	seed.Orders = []orders.Order{{ID: "deadline-one", StoreID: "s1", PurchaseNumber: "8000", Payment: "paid", Status: "awaiting-supply", Preparation: &snapshot}, {ID: "deadline-two", StoreID: "s2", PurchaseNumber: "8001", Payment: "paid", Status: "awaiting-supply", Preparation: &snapshot}, {ID: "legacy-two", StoreID: "s2", Payment: "paid", Status: "awaiting-supply"}}
	if err = repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo, capacityHTTPClock{now: snapshot.ReadyAt.Add(-time.Nanosecond)})
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	path := "/backoffice/orders/supply-overdue"
	commerceHTTP(t, h, "", "", "GET", path, nil, "", 401)
	commerceHTTP(t, h, viewer, "", "GET", path, nil, "", 403)
	commerceHTTP(t, h, seller, "", "GET", path, nil, "", 401)
	var view backoffice.SupplyOverdueView
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", path, nil, "", 200), &view)
	if len(view.Rows) != 0 {
		t.Fatal("early deadline", view)
	}
	repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.ClosedWeekdays = []int{0, 1, 2, 3, 4, 5}
		w.Commerce.Settings.Holidays = []string{"2026-11-29"}
		w.Stores[0].LeadDays = 99
		return nil
	})
	restarted, err := reviewpostgres.New(ctx, db, seed)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := restarted.Snapshot(ctx)
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted, capacityHTTPClock{now: snapshot.ReadyAt})
	admin = login(t, h, "09000000001")
	seller = sellerLogin(t, h, "09120000002")
	for i := 0; i < 3; i++ {
		json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", path, nil, "", 200), &view)
		if len(view.Rows) != 2 || view.MissingDeadlineCount != 1 || !view.Rows[0].Deadline.Equal(snapshot.ReadyAt) || view.Rows[0].CalendarVersion != 3 {
			t.Fatal(view)
		}
	}
	json.Unmarshal(commerceHTTP(t, h, seller, "", "GET", "/seller/orders/supply-overdue", nil, "", 200), &view)
	if len(view.Rows) != 1 || view.Rows[0].StoreID != "s1" || view.MissingDeadlineCount != 0 {
		t.Fatal("scope", view)
	}
	commerceHTTP(t, h, seller, "", "GET", "/seller/orders/supply-overdue?store_id=s2", nil, "", 403)
	request := httptest.NewRequest("GET", path, nil)
	request.Header.Set("Authorization", "Bearer "+admin)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private queue cache")
	}
	for _, private := range []string{"recipient", "phone", "wallet", "receipt"} {
		if strings.Contains(response.Body.String(), `"`+private+`"`) {
			t.Fatal("unneeded private field", private)
		}
	}
	after, _ := restarted.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("monitor mutated persisted state")
	}
}
