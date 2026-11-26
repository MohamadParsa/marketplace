package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func httpSlotQuota(n int) *int { return &n }

type capacityHTTPClock struct{ now time.Time }

func (c capacityHTTPClock) Now() time.Time { return c.now }
func httpCapacitySeed(t *testing.T) backoffice.Workspace {
	w := httpStoreShippingSeed(t)
	w.Commerce.Settings.ClosedWeekdays = []int{}
	for i := range w.Stores {
		w.Stores[i].LeadDays = 0
	}
	m := &w.Commerce.Settings.Shipping[0]
	m.DailyCapacity = httpSlotQuota(1)
	m.DeliveryMode = "slots"
	m.MinDays = 0
	m.MaxDays = 0
	m.DayBasis = shipping.CalendarDays
	m.Schedule = &shipping.Schedule{Windows: []shipping.TimeWindow{{ID: "test-window", Start: "23:00", End: "23:59", Active: true}}, Days: []shipping.WeeklyDay{}}
	for weekday := 0; weekday < 7; weekday++ {
		m.Schedule.Days = append(m.Schedule.Days, shipping.WeeklyDay{Weekday: weekday, Active: true, Slots: []shipping.WindowAssignment{{WindowID: "test-window", Capacity: httpSlotQuota(1)}}})
	}
	return w
}

func TestShippingCapacityPostgresConcurrentLastUnitRestartAndAPIChoices(t *testing.T) {
	ctx := context.Background()
	clock := capacityHTTPClock{now: time.Date(2026, 11, 21, 8, 0, 0, 0, shipping.Tehran())}
	repo, db := cachePostgres(t)
	seed := httpCapacitySeed(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo, clock)
	admin := login(t, h, "09000000001")
	seller := sellerLogin(t, h, "09120000002")
	commerceHTTP(t, h, seller, "", "GET", "/seller/stores/shipping/capacity?store_id=s2&method_id=post", nil, "", 403)
	commerceHTTP(t, h, seller, "", "GET", "/backoffice/stores/shipping/capacity?store_id=s1&method_id=post", nil, "", 401)
	var tokens []string
	var quotes []backoffice.QuoteView
	var date string
	for i, phone := range []string{"09000000003", "09000000004"} {
		token := loginAt(t, h, "storefront", phone)["token"].(string)
		tokens = append(tokens, token)
		var cart backoffice.CartView
		json.Unmarshal(commerceHTTP(t, h, token, "", "GET", "/storefront/cart", nil, "", 200), &cart)
		json.Unmarshal(commerceHTTP(t, h, token, "", "POST", "/storefront/cart", backoffice.CartMutation{Action: "add", OfferID: "offer", Quantity: 1, Version: cart.Version}, fmt.Sprintf("add-%d", i), 200), &cart)
		var address commerce.Address
		json.Unmarshal(commerceHTTP(t, h, token, "", "POST", "/storefront/account/addresses", backoffice.AddressInput{HouseNumber: "12", RecipientSelf: new(bool), RecipientNationalID: "0013547895", Recipient: "گیرنده آزمون ظرفیت", Phone: phone, Province: "تهران", City: "تهران", Street: "نشانی ساختگی", PostalCode: "1234567890"}, fmt.Sprintf("address-%d", i), 200), &address)
		var choices []backoffice.StoreShippingChoices
		json.Unmarshal(commerceHTTP(t, h, token, "", "GET", "/storefront/checkout/shipping-options?address_id="+address.ID, nil, "", 200), &choices)
		if len(choices) != 1 || len(choices[0].Methods) != 1 || len(choices[0].Methods[0].Days) != 6 {
			t.Fatal("availability", choices)
		}
		available := choices[0].Methods[0]
		date = available.Days[0].Date
		last, _ := time.Parse("2006-01-02", date)
		if available.Days[5].Date != last.AddDate(0, 0, 5).Format("2006-01-02") {
			t.Fatal("fixed six dates", available.Days)
		}
		plan := available.Days[0].Slots[0].Plan
		input := backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post"}, PaymentMethodID: "card-transfer", ShippingSelections: map[string]shipping.DeliverySelection{"s1": {Date: date, WindowID: plan.Slot.WindowID}}}
		bad := input
		bad.ShippingSelections = map[string]shipping.DeliverySelection{"s1": {Date: date, WindowID: "forged"}}
		commerceHTTP(t, h, token, "", "POST", "/storefront/checkout/quotes", bad, fmt.Sprintf("forged-%d", i), 400)
		bad.ShippingSelections = map[string]shipping.DeliverySelection{"s1": {Date: last.AddDate(0, 0, 6).Format("2006-01-02"), WindowID: plan.Slot.WindowID}}
		commerceHTTP(t, h, token, "", "POST", "/storefront/checkout/quotes", bad, fmt.Sprintf("outside-%d", i), 400)
		var quote backoffice.QuoteView
		json.Unmarshal(commerceHTTP(t, h, token, "", "POST", "/storefront/checkout/quotes", input, fmt.Sprintf("quote-%d", i), 200), &quote)
		if quote.Shipments[0].Delivery == nil {
			t.Fatal("missing plan snapshot")
		}
		quotes = append(quotes, quote)
	}
	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 2)
	for i := range tokens {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body, _ := json.Marshal(backoffice.AcceptQuoteInput{QuoteID: quotes[i].ID, AcceptedTotalToman: quotes[i].TotalToman})
			r := httptest.NewRequest(http.MethodPost, "/storefront/checkout/purchases", bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", "Bearer "+tokens[i])
			r.Header.Set("Idempotency-Key", fmt.Sprintf("race-capacity-%d", i))
			response := httptest.NewRecorder()
			h.ServeHTTP(response, r)
			results <- response
		}(i)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for response := range results {
		if response.Code == 200 {
			success++
		} else if response.Code == 409 {
			conflict++
		} else {
			t.Fatal(response.Code, response.Body.String())
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal(success, conflict)
	}
	w, _ := repo.Snapshot(ctx)
	if len(w.Commerce.ShippingReservations) != 1 || w.Offers[0].Reserved != 1 {
		t.Fatal("partial/doubled reservation", w.Commerce.ShippingReservations, w.Offers[0])
	}
	restarted, err := reviewpostgres.New(ctx, db, seed)
	if err != nil {
		t.Fatal(err)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted, clock)
	admin = login(t, h2, "09000000001")
	var capacity backoffice.ShippingCapacityView
	json.Unmarshal(commerceHTTP(t, h2, admin, "", "GET", "/backoffice/stores/shipping/capacity?store_id=s1&method_id=post&date="+date, nil, "", 200), &capacity)
	if capacity.Used != 1 || capacity.Held != 1 || capacity.Remaining != 0 || len(capacity.Slots) != 1 || capacity.Slots[0].Remaining != 0 {
		t.Fatal("restart lost capacity", capacity)
	}
	if err := restarted.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Commerce.Quotes {
			if w.Commerce.Quotes[i].Status == "reserved" {
				w.Commerce.Quotes[i].ReservedUntil = clock.Now().Add(-time.Second)
			}
		}
		w.Commerce.ShippingReservations[0].ExpiresAt = clock.Now().Add(-time.Second)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	commerceHTTP(t, h2, admin, "", "GET", "/backoffice/commerce", nil, "", 200)
	w, _ = restarted.Snapshot(ctx)
	if w.Commerce.ShippingReservations[0].Status != "released" || w.Offers[0].Reserved != 0 || len(w.Commerce.Entries) != 0 {
		t.Fatal("expiry not atomic", w.Commerce.ShippingReservations)
	}
}
