package bootstrap_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func TestSellerOffersScopedToUnifiedIdentity(t *testing.T) {
	handler := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	sellerSession := loginAt(t, handler, "seller", "+98 900 000 0002")
	backofficeSession := loginAt(t, handler, "backoffice", "09000000002")
	var sellerUser, backofficeUser identity.User
	for _, account := range []struct {
		session map[string]any
		user    *identity.User
	}{{sellerSession, &sellerUser}, {backofficeSession, &backofficeUser}} {
		raw, err := json.Marshal(account.session["user"])
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, account.user); err != nil {
			t.Fatal(err)
		}
	}
	if sellerUser.ID == "" || sellerUser.ID != backofficeUser.ID || !slices.Equal(sellerUser.StoreIDs, []string{"s1", "s2"}) {
		t.Fatalf("expected one account with two store memberships: seller=%+v backoffice=%+v", sellerUser, backofficeUser)
	}

	call := func(session map[string]any, method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+session["token"].(string))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: got %d, want %d (%s)", method, path, response.Code, want, response.Body)
		}
		return response
	}

	offers := func(session map[string]any, storeID string) []catalog.Offer {
		t.Helper()
		response := call(session, "GET", "/seller/stores/"+storeID+"/offers", "", http.StatusOK)
		var result []catalog.Offer
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result) != 1 || result[0].StoreID != storeID {
			t.Fatalf("expected only offers for %s: %+v", storeID, result)
		}
		return result
	}

	const update = `{"stock":9,"version":0}`
	call(backofficeSession, "GET", "/seller/stores/s1/offers", "", http.StatusUnauthorized)
	call(sellerSession, "GET", "/seller/stores/s4/offers", "", http.StatusForbidden)
	otherSeller := loginAt(t, handler, "seller", "09120000002")
	call(otherSeller, "POST", "/backoffice/commands", `{"action":"user.role","target":"seller-test","value":"super-admin","idempotency_key":"denied"}`, http.StatusUnauthorized)
	offers(otherSeller, "s1")
	call(otherSeller, "GET", "/seller/stores/s2/offers", "", http.StatusForbidden)
	call(otherSeller, "PUT", "/seller/stores/s1/offers/of1/stock", update, http.StatusForbidden)
	call(sellerSession, "PUT", "/seller/stores/s1/offers/of2/stock", update, http.StatusNotFound)
	call(sellerSession, "PUT", "/seller/stores/s2/offers/of2/stock", update, http.StatusForbidden)
	call(sellerSession, "PUT", "/seller/stores/s1/offers/of1/stock", update, http.StatusNoContent)
	call(sellerSession, "PUT", "/seller/stores/s1/offers/of1/stock", update, http.StatusConflict)
	if offer := offers(sellerSession, "s1")[0]; offer.Stock != 9 || offer.Version != 1 {
		t.Fatalf("stock changed incorrectly: %+v", offer)
	}
	if offer := offers(sellerSession, "s2")[0]; offer.Stock != 24 || offer.Version != 0 {
		t.Fatalf("read-only offer changed: %+v", offer)
	}
}
