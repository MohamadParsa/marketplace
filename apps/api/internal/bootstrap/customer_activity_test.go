package bootstrap_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func TestServerBehaviorGuestSearchDedupeAndAdminFunnel(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: "storefront", OperationLogDir: filepath.Join(t.TempDir(), "logs")}, repo)
	defer closeLogHandler(t, h)
	admin := login(t, h, "09000000001")
	operator := login(t, h, "09000000002")
	var guest map[string]string
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", nil, "", 200), &guest)
	mutation := backoffice.CartMutation{Action: "add", OfferID: "offer", Quantity: 1, Version: 1}
	for i := 0; i < 2; i++ {
		commerceHTTP(t, h, "", guest["cart_token"], "POST", "/storefront/cart", mutation, "same-add", 200)
	}
	search := "PRIVATE-SEARCH-09120000001"
	for _, path := range []string{"/storefront/products/product", "/storefront/catalog?q=" + url.QueryEscape(search)} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-Cart-Token", guest["cart_token"])
		r.Header.Set("User-Agent", "Mobile fixture")
		out := httptest.NewRecorder()
		h.ServeHTTP(out, r)
		if out.Code != 200 {
			t.Fatal(out.Code, out.Body.String())
		}
		if out.Header().Get("Set-Cookie") != "" {
			t.Fatal("tracking cookie created")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.(interface{ FlushLogs(context.Context) error }).FlushLogs(ctx); err != nil {
		t.Fatal(err)
	}
	body := commerceHTTP(t, h, admin, "", "POST", "/backoffice/observability/events/query", map[string]any{"log_type": "behavior", "limit": 200}, "", 200)
	for _, private := range []string{search, guest["cart_token"], "09120000001"} {
		if strings.Contains(string(body), private) {
			t.Fatal("private behavior data stored")
		}
	}
	var page operationlog.Page
	json.Unmarshal(body, &page)
	counts := map[string]int{}
	for _, e := range page.Items {
		counts[e.EventName]++
		if e.EventName == "product.viewed" || e.EventName == "search.completed" {
			if !strings.HasPrefix(e.SessionRef, "g-") || e.SubjectID != "" || e.Attributes["device"] != "mobile" {
				t.Fatal("guest identity/device wrong", e)
			}
		}
	}
	if counts["cart.item-added"] != 1 || counts["product.viewed"] != 1 || counts["search.completed"] != 1 {
		t.Fatal(counts)
	}
	result := commerceHTTP(t, h, admin, "", "POST", "/backoffice/observability/funnel/query", map[string]any{"device": "mobile"}, "", 200)
	var funnel backoffice.ActivityFunnel
	json.Unmarshal(result, &funnel)
	if funnel.Environment != "test" || funnel.MatchedEvents != 2 || funnel.UniquePaidPurchases != 0 {
		t.Fatal(string(result))
	}
	commerceHTTP(t, h, operator, "", "POST", "/backoffice/observability/funnel/query", map[string]any{}, "", 403)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/observability/funnel/query", map[string]any{"environment": "production"}, "", 200)
	workspace := commerceHTTP(t, h, admin, "", "GET", "/backoffice/workspace", nil, "", 200)
	if strings.Contains(string(workspace), "activity_origins") {
		t.Fatal("private activity metadata exposed")
	}
}
func TestPostgresBehaviorCartLookupOnEmptyAndValidCart(t *testing.T) {
	repo, _ := cachePostgres(t)
	ctx := context.Background()
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = commerceSeed(); return nil }); err != nil {
		t.Fatal(err)
	}
	if id, err := repo.FindBehaviorCart(ctx, "unknown"); err != nil || id != "" {
		t.Fatal(id, err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	var guest map[string]string
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", nil, "", 200), &guest)
	state, err := repo.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cart := state.Commerce.Carts[0]
	if id, err := repo.FindBehaviorCart(ctx, cart.GuestHash); err != nil || id != cart.ID {
		t.Fatal(id, err)
	}
}
